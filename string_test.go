package tools

import (
	"log"
	"math/rand"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestName(t *testing.T) {
	//u := &User{Id: 23132, Username: "foo", Password: "bar", Identity: "324213", QrCodes: nil, DenKey: "ansssss", TalkingKey: "qwesad"}
	////Show(u)
	//s := Save(u)
	//log.Println(s)
	//m := make(map[string]interface{})
	//Lock()
	for i := 0; i < 10; i++ {
		go func(key int) {
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			LockFunc("check", func() {
				log.Println("Checking:", key)
			})
		}(i)
	}
	time.Sleep(50 * time.Second)
	//s := Make("")
	//s.Marshal(u)
	//log.Println(s)
}

func findStr(name string) {
	log.Println("findStr")
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

func TestStr(t *testing.T) {
	s := Make("（林婕琼）")
	log.Println(Make("12345").FormatterNum())
	log.Println(s.GetRune("林琼"))
}
