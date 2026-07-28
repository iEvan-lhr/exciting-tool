package tools

import (
	"net/http"
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	type App struct {
		AppName  string `json:"app_name"`
		Password string `json:"password"`
	}
	a := App{}
	request, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"app_name":"exciting-tool","password":"secret"}`))
	if err != nil {
		t.Fatal(err)
	}
	MarshalReq([]any{request}, &a)
	if a.AppName != "exciting-tool" || a.Password != "secret" {
		t.Fatalf("MarshalReq decoded %+v", a)
	}
}

func Success(str string) string {
	if str == "tempSuccess" {
		return "SSSS"
	}
	return "OOOO"
}

func Fail(str string) string {
	if str == "tempFail" {
		return "Fail"
	}
	return "OOOO"
}
