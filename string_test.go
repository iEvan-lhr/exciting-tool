package tools

import (
	"log"
	"os"
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
	////s.Marshal(&User{UserName: "foo", Password: "bar", Order: 3.23})
	////log.Println(s)
	//log.Println(Make("").Save(User{Id: 23132, Username: "foo", Password: "bar", Identity: "123sakdjwe", QrCode: "982j32", DenKey: "ansssss", TalkingKey: "qwesad"}))
	//s := Make("99999")
	//s.ReplaceLastStr(1, "888")
	var users []User
	user := User{Id: 23132, Username: "foo", Password: "bar", Identity: "123sakdjwe", QrCode: "982j32", DenKey: "ansssss", TalkingKey: "qwesad"}
	for i := 0; i < 560; i++ {
		users = append(users, user)
	}
	save := Save(users)
	var buf []byte
	for i := range save {
		buf = append(buf, save[i].buf...)
	}
	err := os.WriteFile("testing.txt", buf, 0666)
	Error(err)
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
