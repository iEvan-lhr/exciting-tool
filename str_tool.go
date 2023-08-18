package tools

import (
	"log"
	"reflect"
	"unicode/utf8"
)

const EndMessage = "----------END----------"
const Select = "select * from "

func inTF(i, j int) bool {
	return i == j
}

func marshalStruct(model any) (result []*String) {
	values, typ := returnValAndTyp(model)
	switch values.Kind() {
	case reflect.Struct:
		return generateModel(values, typ)
	case reflect.Slice:
		return generateModels(values)
	}
	return nil
}

func generateModel(values reflect.Value, types reflect.Type) (result []*String) {
	s := Make("insert into ")
	s.cutHumpMessage(values.String())
	tags := Make(" (")
	vars := Make(" values(")
	for j := 0; j < types.NumField(); j++ {
		switch types.Field(j).Tag.Get("marshal") {
		case "off":
		case "auto_insert":
			tags.Append("`", humpName(types.Field(j).Name), "`,")
			vars.appendAny("NULL,")
		default:
			tags.Append("`", humpName(types.Field(j).Name), "`,")
			vars.Append("'", righteousCharacter(Make(values.Field(j).Interface())), "',")
		}
	}
	tags.ReplaceLastStr(1, ")")
	vars.ReplaceLastStr(1, ")")
	s.Append(tags, vars)
	result = append(result, s)
	return
}

func generateModels(values reflect.Value) (result []*String) {
	if !(values.Len() > 0) {
		return
	}
	head, lens := generateHead(values.Index(0).Interface())
	s := Make(head)
	for i := 0; i < values.Len(); i++ {
		v, t := returnValAndTyp(values.Index(i).Interface())
		if i != 0 && i%200 == 0 {
			s.ReplaceLastStr(1, ";\n")
			result = append(result, s)
			s = Make(head)
		}
		vars := Make("(")
		for j := 0; j < lens; j++ {
			switch t.Field(j).Tag.Get("marshal") {
			case "off":
			case "auto_insert":
				vars.appendAny("NULL,")
			default:
				vars.Append("'", righteousCharacter(Make(v.Field(j).Interface())), "',")
			}
		}
		vars.ReplaceLastStr(1, "),")
		s.appendAny(vars)
	}
	s.ReplaceLastStr(1, ";")
	result = append(result, s)
	return
}

func generateHead(model any) (*String, int) {
	values, typ := returnValAndTyp(model)
	s := Make("insert into ")
	s.cutHumpMessage(values.String())
	tags := Make(" (")
	for j := 0; j < typ.NumField(); j++ {
		switch typ.Field(j).Tag.Get("marshal") {
		case "off":
		case "auto_insert":
			tags.Append("`", humpName(typ.Field(j).Name), "`,")
		default:
			tags.Append("`", humpName(typ.Field(j).Name), "`,")
		}
	}
	tags.ReplaceLastStr(1, ")")
	s.Append(tags, " values")
	return s, typ.NumField()
}

func (s *String) queryStruct(model any) {
	values, typ := returnValAndTyp(model)
	s.appendAny(Select)
	s.cutHumpMessage(values.String())
	var where byte
	for j := 0; j < typ.NumField(); j++ {
		if !values.Field(j).IsZero() {
			if where == 0 {
				s.appendAny(" where ")
				where++
			} else {
				s.appendAny(" and ")
			}
			switch values.Field(j).Kind() {
			case reflect.Slice:
			default:
				s.Append(humpName(typ.Field(j).Name), "=", "'", righteousCharacter(Make(values.Field(j).Interface())), "'")
			}
		}
	}
}

func (s *String) checkStruct(model any) {
	values, typ := returnValAndTyp(model)
	s.appendAny(Select)
	s.cutHumpMessage(values.String())
	var where byte
	for j := 0; j < typ.NumField(); j++ {
		if !values.Field(j).IsZero() && typ.Field(j).Tag.Get("marshal") == "check" {
			if where == 0 {
				s.appendAny(" where ")
				where++
			} else {
				s.appendAny(" and ")
			}
			switch values.Field(j).Kind() {
			case reflect.Slice:
			default:
				s.Append(humpName(typ.Field(j).Name), "=", "'", righteousCharacter(Make(values.Field(j).Interface())), "'")
			}
		}
	}
}

func (s *String) cutStructMessage(sm string) {
	sms := Make(sm)
	split := sms.Split(".")
	sms.coverWrite(split[len(split)-1])
	s.Append("\n", "----------", sms.Split(" ")[0], "----------", "\n")
}

func (s *String) cutHumpMessage(hump string) {
	sms := Make(hump)
	split := sms.Split(".")
	sms.coverWrite(split[len(split)-1])
	sms.coverWrite(humpName(sms.Split(" ")[0]))
	s.appendAny(sms)
}

// AppendSpilt  拼接字符串后返回String
// use AppendSpiltLR(",",24,23,22,21,11)
// get "24,23,22,21,11"
// or AppendSpiltLR(",",[]any...])
func (s *String) AppendSpilt(join ...any) *String {
	var split = &String{}
	for i := range join {
		if i == 0 {
			split.Append(join[i])
		} else if i == len(join)-1 {
			s.appendAny(join[i])
		} else {
			s.appendAny(join[i])
			s.Append(split)
		}

	}
	return s
}

// AppendSpiltLR  拼接字符串后返回String
// use AppendSpiltLR(",","[","]",24,23,22,21,11)
// get "[24,23,22,21,11]"
// or AppendSpiltLR(",","[","]",[]any...])
// use Make("insert userinfo values(").AppendSpiltLR(",", "'", "'", 4, 5, 6, 7, 8, 9, 10).Append(")")
// get insert userinfo values('4','5','6','7','8','9','10')
func (s *String) AppendSpiltLR(join ...any) *String {
	var split, l, r = &String{}, &String{}, &String{}
	if len(join) < 3 {
		panic("Add Lens<3")
	}
	split.appendAny(join[0])
	l.appendAny(join[1])
	r.appendAny(join[2])
	for i := 3; i < len(join)-1; i++ {
		s.appendAny(l)
		s.appendAny(join[i])
		s.appendAny(r)
		s.Append(split)
	}
	s.appendAny(l)
	s.appendAny(join[len(join)-1])
	s.appendAny(r)
	return s
}

// checkBytes 比较两个byte切片的值
func checkBytes(s, str []byte) bool {
	if inTF(len(s), len(str)) {
		for i, v := range str {
			if s[i] != v {
				return false
			}
		}
		return true
	}
	return false
}

// RunesToBytes  Runes转bytes
func runesToBytes(rune []rune) []byte {
	size := 0
	for _, r := range rune {
		size += utf8.RuneLen(r)
	}
	bs := make([]byte, size)
	count := 0
	for _, r := range rune {
		count += utf8.EncodeRune(bs[count:], r)
	}
	return bs
}
func (s *String) Marshal(model any) {
	values, types := returnValAndTyp(model)
	s.cutStructMessage(values.String())
	for j := 0; j < types.NumField(); j++ {
		if types.Field(j).Tag.Get("marshal") != "off" {
			s.Append(types.Field(j).Name, ":", values.Field(j).Interface(), "\n")
		}
	}
	s.appendAny(EndMessage)
}

func Show(show any) {
	s := &String{}
	values, types := returnValAndTyp(show)
	s.cutStructMessage(values.String())
	for j := 0; j < types.NumField(); j++ {
		if types.Field(j).Tag.Get("marshal") != "off" {
			s.Append(types.Field(j).Name, ":", values.Field(j).Interface(), "\n")
		}
	}
	s.appendAny(EndMessage)
	log.Println(s)
}

func returnValAndTyp(model any) (values reflect.Value, types reflect.Type) {
	switch reflect.ValueOf(model).Kind() {
	case reflect.Struct, reflect.Slice:
		values = reflect.ValueOf(model)
		types = reflect.TypeOf(model)
	case reflect.Pointer:
		values = reflect.ValueOf(model).Elem()
		types = reflect.TypeOf(model).Elem()
	case reflect.Map:

	}
	return
}

func MarshalMap(model any) map[string]string {
	modelMap := make(map[string]string)
	var values reflect.Value
	var types reflect.Type
	switch reflect.ValueOf(model).Kind() {
	case reflect.Struct:
		values = reflect.ValueOf(model)
		types = reflect.TypeOf(model)
	case reflect.Pointer:
		values = reflect.ValueOf(model).Elem()
		types = reflect.TypeOf(model).Elem()
	}
	modelMap["StructName"] = cutStructMessage(values.String())
	for j := 0; j < types.NumField(); j++ {
		if types.Field(j).Tag.Get("marshal") != "off" {
			modelMap[types.Field(j).Name] = Make(values.Field(j).Interface()).string()
		}
	}
	return modelMap
}

// IsNumber 用来检测字符串是否为数字
func (s *String) IsNumber() bool {
	_, err := s.Atoi()
	if err != nil {
		return false
	}
	return true
}

// FormatterNum 格式化输出字符串
func (s *String) FormatterNum() (bool, string) {
	if !s.IsNumber() {
		return false, s.string()
	}
	result := Make()
	m := 1
	for _, v := range s.buf {
		if s.Len() != m && (s.Len()-m)%3 == 0 {
			result.Append(v, ",")
		} else {
			result.appendAny(v)
		}
		m++
	}
	return true, result.string()
}
