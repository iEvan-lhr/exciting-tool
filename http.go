package tools

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/iEvan-lhr/exciting-tool/httpx"
)

var defaultHTTPClient = &http.Client{Timeout: 30 * time.Second}

// NewHTTPClient creates the preferred error-returning HTTP client.
func NewHTTPClient(options ...httpx.Option) *httpx.Client {
	return httpx.New(options...)
}

// Do executes a GET request when body is omitted, otherwise a POST request.
// Deprecated: Use httpx.Client or NewHTTPClient.
func Do(url string, args ...interface{}) *String {
	if len(args) == 0 {
		return get(url, nil)
	}
	return post(url, Make(args[0]).String(), nil)
}

// DoReq executes a request with one or more clients and concatenates responses.
// Deprecated: Use httpx.Client.Do.
func DoReq(r *http.Request, clients ...*http.Client) *String {
	if r == nil {
		return Make()
	}
	if len(clients) == 0 {
		return executeRequest(defaultHTTPClient, r)
	}

	body, err := requestBody(r)
	ExecError(err)
	result := Make()
	for _, client := range clients {
		if client == nil {
			client = defaultHTTPClient
		}
		request := r.Clone(r.Context())
		if body != nil {
			request.Body = io.NopCloser(bytes.NewReader(body))
			request.ContentLength = int64(len(body))
		}
		result.Append(executeRequest(client, request))
	}
	return result
}

// DoUseHeader executes a GET or POST request with custom headers.
// Deprecated: Use httpx.Client.Do.
func DoUseHeader(url string, header http.Header, args ...interface{}) *String {
	if len(args) == 0 {
		return get(url, header)
	}
	return post(url, Make(args[0]).String(), header)
}

// UnMarshal parses a request body into v.
// Deprecated: If iEvan-lhr/worker constructs the request, use MarshalReq.
func UnMarshal(r *http.Request, v interface{}) interface{} {
	if r == nil || r.Body == nil || v == nil {
		return v
	}
	data, err := io.ReadAll(r.Body)
	ExecError(err)
	Unmarshal(data, v)
	return v
}

// MarshalReq parses the first request in r into v.
func MarshalReq(r []any, v interface{}) interface{} {
	if len(r) == 0 || v == nil {
		return v
	}
	request, ok := r[0].(*http.Request)
	if !ok {
		return v
	}
	return UnMarshal(request, v)
}

func get(url string, header http.Header) *String {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	ExecError(err)
	setRequestHeader(request, header)
	return executeRequest(defaultHTTPClient, request)
}

func post(url, body string, header http.Header) *String {
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	ExecError(err)
	setRequestHeader(request, header)
	return executeRequest(defaultHTTPClient, request)
}

func executeRequest(client *http.Client, request *http.Request) *String {
	response, err := client.Do(request)
	ExecError(err)
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	ExecError(err)
	return BytesString(data)
}

func requestBody(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}
	if request.GetBody != nil {
		body, err := request.GetBody()
		if err != nil {
			return nil, err
		}
		defer body.Close()
		return io.ReadAll(body)
	}

	data, err := io.ReadAll(request.Body)
	if closeErr := request.Body.Close(); err == nil {
		err = closeErr
	}
	request.Body = io.NopCloser(bytes.NewReader(data))
	return data, err
}

func setRequestHeader(request *http.Request, header http.Header) {
	if header == nil {
		request.Header = headerPublic()
		return
	}
	request.Header = header.Clone()
}

func headerPublic() http.Header {
	header := http.Header{}
	header.Set("Accept", "*/*")
	header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	header.Set("Content-Type", "application/json")
	header.Set("User-Agent", "exciting-tool")
	return header
}
