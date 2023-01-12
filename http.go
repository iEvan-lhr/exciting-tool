package tools

import (
	"io"
	"net/http"
	"reflect"
	"strings"
)

// Do 用于执行标准的请求方法,默认设置的header如下
// header.Set("Accept", "*/*")
// header.Set("Accept-Language", "zh-CN,zh;q=0.9")
// header.Set("Connection", "keep-alive")
// header.Set("Content-Type", "application/json")
// header.Set("User-Agent", "PostmanRuntime/7.28.4")
func Do(url string, args ...interface{}) *String {
	if len(args) <= 0 {
		return get(url, nil)
	} else {
		return post(url, args[0].(string), nil)
	}
}

// DoReq 用于执行需要自定义client的请求方法，例如使用https 可以设置信任证书等等 入参可选传入client 传入多个将拼接多次的执行结果
func DoReq(r *http.Request, client ...*http.Client) *String {
	if len(client) > 0 {
		result := Make()
		for i := range client {
			result.appendAny(ReturnValueByTwo(io.ReadAll(ReturnValueByTwo(client[i].Do(r)).(*http.Response).Body)))
		}
		return result
	}
	return Make(ReturnValueByTwo(io.ReadAll(ReturnValueByTwo((&http.Client{}).Do(r)).(*http.Response).Body)))
}

// DoUseHeader 用于执行自定义header的请求方法 入参为请求地址、header、body及其他参数
// 允许header为nil 将会使用默认的header
func DoUseHeader(url string, header http.Header, args ...interface{}) *String {
	if len(args) <= 0 {
		return get(url, header)
	} else {
		return post(url, args[0].(string), header)
	}
}

// UnMarshal 用于从request中解析参数
// Deprecated: 若使用iEvan-lhr/worker 构建request , 推荐使用 MarshalReq.
func UnMarshal(r *http.Request, v interface{}) interface{} {
	Unmarshal(ReturnValueByTwo(io.ReadAll(r.Body)), reflect.ValueOf(v).Interface())
	return v
}

// MarshalReq 用于从request中解析参数
func MarshalReq(r []interface{}, v interface{}) interface{} {
	Unmarshal(ReturnValueByTwo(io.ReadAll(r[0].(*http.Request).Body)), reflect.ValueOf(v).Interface())
	return v
}

// get 底层方法 用于发送get请求
func get(url string, header http.Header) *String {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	if header == nil {
		request.Header = headerPublic()
	} else {
		request.Header = header
	}
	return Make(ReturnValueByTwo(io.ReadAll(ReturnValueByTwo((&http.Client{}).Do(request)).(*http.Response).Body)))
}

// post 底层方法 用于发送post请求
func post(url, body string, header http.Header) *String {
	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	if header == nil {
		request.Header = headerPublic()
	} else {
		request.Header = header
	}
	return Make(ReturnValueByTwo(io.ReadAll(ReturnValueByTwo((&http.Client{}).Do(request)).(*http.Response).Body)))
}

// headerPublic 设置标准请求头
func headerPublic() http.Header {
	header := http.Header{}
	header.Set("Accept", "*/*")
	header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	header.Set("Connection", "keep-alive")
	header.Set("Content-Type", "application/json")
	header.Set("User-Agent", "PostmanRuntime/7.28.4")
	return header
}
