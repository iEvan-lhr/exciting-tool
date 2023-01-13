package tools

import (
	"bytes"
	"reflect"
	"time"
)

// humpName 格式化驼峰命名
func humpName(buf string) (ans []byte) {
	if len(buf) > 0 {
		for i := range buf {
			if buf[i] < 97 {
				if i == 0 {
					ans = append(ans, buf[0]+32)
				} else {
					ans = append(ans, sli)
					ans = append(ans, buf[i]+32)
				}
			} else {
				ans = append(ans, buf[i])
			}
		}
	}
	return
}

func righteousCharacter(s *String) *String {
	var runes []rune
	for _, v := range bytes.Runes(s.buf) {
		if v == '\'' || v == '`' {
			runes = append(runes, append([]rune(`\`), v)...)
		} else {
			runes = append(runes, v)
		}
	}
	s.runes = runes
	s.buf = runesToBytes(s.Runes())
	return s
}

func marshalTable(model any) *String {
	values, typ := returnValAndTyp(model)
	if values.Kind() == reflect.Struct {
		return saveTable(values, typ)
	} else {
		panic("unsupported type for marshalTable : has to be struct")
	}
}

func saveTable(values reflect.Value, types reflect.Type) *String {
	s := Make("CREATE TABLE `")
	s.cutHumpMessage(values.String())
	s.appendAny("` (\n")
	for j := 0; j < types.NumField(); j++ {
		s.Append(humpName(types.Field(j).Name), "   ", returnType(values.Field(j)))
		switch types.Field(j).Tag.Get("marshal") {
		case "pro":
			s.Append("` primary key")
		case "default":
			s.appendAny(" " + types.Field(j).Tag.Get("default") + "\n")
		case "":

		}
		s.appendAny(",\n")
	}
	s.ReplaceLastStr(2, "\n)")
	//result = append(result, s)
	return s
}

func returnType(typ reflect.Value) string {
	switch typ.Kind() {
	case 24:
		return "varchar(200)"
	case 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
		return "int"
	case 13, 14:
		return "float"
	case 25:
		switch typ.Interface().(type) {
		case time.Time:
			return "Data"
		}
	}
	return ""
}
