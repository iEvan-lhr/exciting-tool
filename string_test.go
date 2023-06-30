package tools

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestName(t *testing.T) {
	u := &User{Id: 23132, Username: "foo", Password: "bar", Identity: "324213", QrCodes: nil, DenKey: "ansssss", TalkingKey: "qwesad"}
	Show(u)
	//s := Make("")
	//s.Marshal(u)
	//log.Println(s)
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
	QrCodes    []int  `json:"qr_code"`
	DenKey     string `json:"den_key"`
	TalkingKey string `json:"talking_key"`
}

//func (u *User) String() string {
//	return u.Username + ":" + u.Password
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
