package tools

import (
	"log"
	"testing"
)

func TestError(t *testing.T) {
	type App struct {
		AppName  string `json:"app_name"`
		Password string `json:"password"`
	}
	//marshal, _ := json.Marshal(App{
	//	AppName:  "追云鹿",
	//	Password: "ZXC000",
	//})
	a := App{}

	log.Println(UnMarshal(nil, &a))
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
