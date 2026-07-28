package tools

import (
	"bytes"
	"reflect"
	"time"
	"unicode"
)

// humpName 格式化驼峰命名
func humpName(buf string) (ans []byte) {
	runes := []rune(buf)
	for i, current := range runes {
		if unicode.IsUpper(current) {
			if i > 0 {
				previous := runes[i-1]
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if unicode.IsLower(previous) || unicode.IsDigit(previous) || nextIsLower {
					ans = append(ans, sli)
				}
			}
			current = unicode.ToLower(current)
		}
		ans = append(ans, string(current)...)
	}
	return
}

func righteousCharacter(s *String) *String {
	s.buf = bytes.ReplaceAll(s.buf, []byte("'"), []byte("''"))
	s.runes = nil
	return s
}

func marshalTable(model any) *String {
	values, typ := returnValAndTyp(model)
	if values.IsValid() && values.Kind() == reflect.Struct {
		return saveTable(values, typ)
	} else {
		panic("unsupported type for marshalTable : has to be struct")
	}
}

func saveTable(values reflect.Value, types reflect.Type) *String {
	s := Make("CREATE TABLE `")
	s.Append(humpName(types.Name()))
	s.appendAny("` (\n")
	for j := 0; j < types.NumField(); j++ {
		field := types.Field(j)
		if field.PkgPath != "" || field.Tag.Get("marshal") == "off" {
			continue
		}
		s.Append("`", humpName(field.Name), "` ", returnType(values.Field(j)))
		switch types.Field(j).Tag.Get("marshal") {
		case "pro":
			s.Append(" primary key")
		case "default":
			s.Append(" default ", types.Field(j).Tag.Get("default"))
		case "":

		}
		s.appendAny(",\n")
	}
	s.ReplaceLastStr(2, "\n)")
	return s
}

func returnType(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.String:
		return "varchar(200)"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "boolean"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	case reflect.Slice:
		if typ.Type().Elem().Kind() == reflect.Uint8 {
			return "blob"
		}
	}
	return "text"
}
