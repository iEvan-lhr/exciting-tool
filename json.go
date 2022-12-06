package tools

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
)

// TransHtmlJson 公共方法 处理Go原生json不会替换html字符的问题
func transHtmlJson(data []byte) []byte {
	data = bytes.Replace(data, []byte("\\u0026"), []byte("&"), -1)
	data = bytes.Replace(data, []byte("\\u003c"), []byte("<"), -1)
	data = bytes.Replace(data, []byte("\\u003e"), []byte(">"), -1)
	return data
}

// Unmarshal 公共方法 解析数据至空模板
func Unmarshal(v interface{}, str interface{}) {
	var s []byte
	switch v.(type) {
	case string:
		s = transHtmlJson([]byte(v.(string)))
	case []byte:
		s = transHtmlJson(v.([]byte))
	default:
		s = transHtmlJson(ReturnValueByTwo(json.Marshal(v)).([]byte))
	}
	switch str.(type) {
	case string:
		reflect.ValueOf(str).Elem().Set(reflect.ValueOf(string(s)))
	case reflect.Value:

	default:
		ExecError(json.Unmarshal(s, &str))
	}
}

func UnmarshalByOriginal(v interface{}, str interface{}) {
	switch v.(type) {
	case string:
		ExecError(json.Unmarshal([]byte(v.(string)), &str))
	case []byte:
		ExecError(json.Unmarshal(v.([]byte), &str))
	default:
		ExecError(json.Unmarshal(ReturnValueByTwo(json.Marshal(v)).([]byte), &str))
	}
}

func UMarshal(v, str interface{}) {
	var marshal []byte
	var m map[string]any
	switch v.(type) {
	case []byte:
		marshal = v.([]byte)
		eatError(json.Unmarshal(marshal, &m))
	case string:
		marshal = []byte(v.(string))
		eatError(json.Unmarshal(marshal, &m))
	default:
		marshal, m = marshalMap(v)
		eatError(json.Unmarshal(marshal, &m))
	}
	eatError(json.Unmarshal(marshal, str))
	values, typ := returnValAndTyp(str)
	if typ.Kind() == reflect.Map {
		values.Set(reflect.ValueOf(m))
		return
	}
	for j := 0; j < typ.NumField(); j++ {
		switch values.Field(j).Interface().(type) {
		case *String:
			flo := m[typ.Field(j).Tag.Get("json")].(float64)
			if float64(int(flo)) == flo {
				values.Field(j).Set(reflect.ValueOf(Make(int(flo))))
			} else {
				values.Field(j).Set(reflect.ValueOf(Make(flo)))
			}
		}
	}
	return
}

func Marshal(v interface{}) []byte {
	values, typ := returnValAndTyp(v)
	m := make(map[string]string)
	for j := 0; j < typ.NumField(); j++ {
		if !values.Field(j).IsZero() {
			m[typ.Field(j).Tag.Get("json")] = Make(values.Field(j).Interface()).string()
		}
	}
	return ReturnValue(json.Marshal(m)).([]byte)
}

func marshalMap(v interface{}) ([]byte, map[string]any) {
	values, typ := returnValAndTyp(v)
	m := make(map[string]any)
	for j := 0; j < typ.NumField(); j++ {
		if values.Field(j).Kind() != reflect.Slice && values.Field(j).Kind() != reflect.Map && !values.Field(j).IsZero() {
			switch values.Field(j).Interface().(type) {
			case *String:
				m[typ.Field(j).Tag.Get("json")] = Make(values.Field(j).Interface())
			default:
				m[typ.Field(j).Tag.Get("json")] = values.Field(j).Interface()
			}

		}
	}
	log.Println(m)
	return ReturnValue(json.Marshal(m)).([]byte), m
}
