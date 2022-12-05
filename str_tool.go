package tools

import (
	"reflect"
	"time"
	"unicode/utf8"
)

const timeLayout = "2006-01-02 15:04:05"

func inTF(i, j int) bool {
	return i == j
}

func (s *String) strTime(t time.Time) {
	const bufSize = 64
	var b []byte
	var buf [bufSize]byte
	b = buf[:0]
	s.appendAny(t.AppendFormat(b, timeLayout))
}

func (s *String) marshalStruct(structs ...any) {
	for i := range structs {
		if reflect.ValueOf(structs[i]).Kind() == 25 {
			values := reflect.ValueOf(structs[i])
			types := reflect.TypeOf(structs[i])
			s.cutStructMessage(values.String())
			for j := 0; j < types.NumField(); j++ {
				s.Append(types.Field(j).Name, ":", values.Field(j).Interface(), "\n")
			}
		}
	}
}

func (s *String) cutStructMessage(sm string) {
	sms := Make(sm)
	split := sms.Split(".")
	sms.coverWrite(split[len(split)-1])
	s.Append(sms.Split(" ")[0], ":", "\n")
}

// AppendSpilt  拼接字符串后返回String
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

// appendKind  拼接字符串
func (s *String) appendKind(join any) int {
	kind := reflect.ValueOf(join).Kind()
	switch kind {
	case 24:
		return ReturnValue(s.writeString(join.(string))).(int)
	case 8:
		ReturnValue(s.WriteByte(join.(byte)))
		return 1
	case 2:
		return appendInt(join.(int), &s.buf)
	case 3:
		return appendInt(int(join.(int8)), &s.buf)
	case 4:
		return appendInt(int(join.(int16)), &s.buf)
	case 5:
		return appendInt(int(join.(int32)), &s.buf)
	case 6:
		return appendInt(int(join.(int64)), &s.buf)
	case 7:
		return appendUint64(uint64(join.(uint)), &s.buf)
	case 9:
		return appendUint64(uint64(join.(uint16)), &s.buf)
	case 10:
		return appendUint64(uint64(join.(uint32)), &s.buf)
	case 11:
		return appendUint64(join.(uint64), &s.buf)
	case 13:
		l1 := s.Len()
		genericFtoa(&s.buf, float64(join.(float32)), 'f', 2, 32)
		return s.Len() - l1
	case 14:
		l1 := s.Len()
		genericFtoa(&s.buf, join.(float64), 'f', 2, 32)
		return s.Len() - l1
	case 1:
		if join.(bool) {
			return ReturnValue(s.writeString(TRUE)).(int)
		} else {
			return ReturnValue(s.writeString(FALSE)).(int)
		}
	case 22:
		return ReturnValue(s.writeString(reflect.ValueOf(join).MethodByName("String").Call(nil)[0].String())).(int)

	default:
		switch join.(type) {
		case time.Time:
			s.strTime(join.(time.Time))
		case *String:
			return ReturnValue(s.Write(join.(*String).buf)).(int)
		case []byte:
			return ReturnValue(s.Write(join.([]byte))).(int)
		default:
			panic("unsupported join type")
		}
	}
	return -1
}

func (s *String) Marshal(model any) {
	if reflect.ValueOf(model).Kind() == 25 {
		values := reflect.ValueOf(model)
		types := reflect.TypeOf(model)
		s.cutStructMessage(values.String())
		for j := 0; j < types.NumField(); j++ {
			s.Append(types.Field(j).Name, ":", values.Field(j).Interface(), "\n")
		}
	}
}
func MarshalMap(model any) map[string]string {
	modelMap := make(map[string]string)
	if reflect.ValueOf(model).Kind() == 25 {
		values := reflect.ValueOf(model)
		types := reflect.TypeOf(model)
		modelMap["StructName"] = cutStructMessage(values.String())
		for j := 0; j < types.NumField(); j++ {
			modelMap[types.Field(j).Name] = Make(values.Field(j).Interface()).string()
		}
	}
	return modelMap
}
