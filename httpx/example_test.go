package httpx_test

import (
	"io"
	"time"

	"github.com/iEvan-lhr/exciting-tool/httpx"
)

func ExampleNew() {
	client := httpx.New(
		httpx.WithTimeout(10*time.Second),
		httpx.WithMaxBodyBytes(2<<20),
	)
	_ = client
}

func ExampleStreamResponse_LimitBody() {
	response := &httpx.StreamResponse{
		Body: io.NopCloser(&zeroReader{}),
	}
	response.LimitBody(1024)
	defer response.Close()
}

type zeroReader struct{}

func (*zeroReader) Read([]byte) (int, error) {
	return 0, io.EOF
}
