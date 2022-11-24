package tools

import (
	"io"
	"net/http"
	"strings"
)

func Do(url string, args ...interface{}) *String {
	if len(args) <= 0 {
		return get(url)
	} else {
		return post(url, args[0].(string))
	}
}
func UnMarshal(r *http.Request, v interface{}) interface{} {
	Unmarshal(ReturnValueByTwo(io.ReadAll(r.Body)), v)
	return v
}

func get(url string) *String {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	request.Header = headerPublic()
	return Make(ReturnValueByTwo(io.ReadAll(ReturnValueByTwo((&http.Client{}).Do(request)).(*http.Response).Body)))
}

func post(url, body string) *String {
	request, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	request.Header = headerPublic()
	return Make(ReturnValueByTwo(io.ReadAll(ReturnValueByTwo((&http.Client{}).Do(request)).(*http.Response).Body)))
}
func headerPublic() http.Header {
	header := http.Header{}
	header.Set("Accept", "*/*")
	header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	header.Set("Connection", "keep-alive")
	header.Set("Content-Type", "application/json")
	header.Set("User-Agent", "PostmanRuntime/7.28.4")
	return header
}
