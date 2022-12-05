package tools

import (
	"log"
	"testing"
)

func TestError(t *testing.T) {
	temp := &ParseError{
		values: nil,
		err:    nil,
		isErr:  true,
	}
	temp1 := *temp
	temp1.isErr = false
	UnMarshal(nil, temp)
	log.Println(temp.isErr, temp1.isErr)
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
