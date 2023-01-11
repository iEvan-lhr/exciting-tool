package tools

import (
	"log"
	"reflect"
	"testing"
	"unsafe"
)

func TestName(t *testing.T) {
	u := &User{Id: 23132, Username: "foo", Password: "bar", Identity: "324213", QrCode: "982j32", DenKey: "ansssss", TalkingKey: "qwesad"}
	//log.Println(Make(u))
	//var app []User
	//for i := 0; i < 10000; i++ {
	//	app = append(app, *u)
	//}
	//Send(0, app)
	log.Println(Make(u))
	//u1:=UserHead(u)
	//log.Println(UserHead(u))
	//b := StructToBytes(u)
	//u.Username = ":::::::"
	//n1 := &User{}
	//n1 = BytesToStruct(b)
	//log.Println(n1)
}

func Send(i int, data []User) {
	if i < 1000 {
		Send(i+1, data)
	}
}

type User struct {
	Id         int    `json:"id" marshal:"auto_insert"`
	Username   string `json:"username"`
	Password   string `json:"password" marshal:"off"`
	Identity   string `json:"identity"`
	QrCode     string `json:"qr_code"`
	DenKey     string `json:"den_key"`
	TalkingKey string `json:"talking_key"`
}

//func (u *User) String() string {
//	return u.Username + ":" + u.Password
//}

//func (u User) setUsername(username string) {
//	u.Username = username
//}
//
//func (u User) setPassword(username string) {
//	u.Password = username
//}

func StructToBytes(model *User) []byte {
	var x reflect.SliceHeader
	x.Len = int(unsafe.Sizeof(model))
	x.Cap = x.Len
	x.Data = uintptr(unsafe.Pointer(model))
	return *(*[]byte)(unsafe.Pointer(&x))
}

func BytesToStruct(data []byte) *User {
	return (*User)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&data)).Data))
}
