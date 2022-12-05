package tools

import (
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	log.Println(Make(time.Now()))
	//
	////for i, i2 := range MarshalMap(User{UserName: "foo", Password: "bar", Order: 3.23}) {
	////	log.Println(i, i2)
	////}
	////log.Println(reflect.ValueOf(&User{UserName: "foo", Password: "bar", Order: 3.23}).MethodByName("String").IsValid())
	s := Make([]any{&User{Username: "foo", Password: "bar"}, &User{Username: "boo", Password: "bar"}, &User{Username: "foo", Password: "coo"}})
	////s.Marshal(&User{UserName: "foo", Password: "bar", Order: 3.23})
	////log.Println(s)
	//log.Println(Make("").Save(User{Id: 23132, Username: "foo", Password: "bar", Identity: "123sakdjwe", QrCode: "982j32", DenKey: "ansssss", TalkingKey: "qwesad"}))
	log.Println(s)
}

type User struct {
	Id         int    `json:"id" marshal:"auto_insert"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Identity   string `json:"identity"`
	QrCode     string `json:"qr_code"`
	DenKey     string `json:"den_key"`
	TalkingKey string `json:"talking_key"`
}
