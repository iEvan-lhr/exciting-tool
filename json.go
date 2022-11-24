package tools

import (
	"bytes"
	"encoding/json"
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
