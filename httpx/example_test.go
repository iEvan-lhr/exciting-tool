package httpx_test

import (
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
