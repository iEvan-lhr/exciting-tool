package tools

import (
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	log.Println(Make(time.Now()))
	type User struct {
		UserName string
		Password string
		Order    float64
	}
	for i, i2 := range MarshalMap(User{UserName: "foo", Password: "bar", Order: 3.23}) {
		log.Println(i, i2)
	}
}
