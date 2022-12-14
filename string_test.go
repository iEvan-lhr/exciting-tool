package tools

import (
	"log"
	"testing"
)

func TestName(t *testing.T) {
	u := User{Id: 23132, Username: "foo", Password: "bar", Identity: "324213", QrCode: "982j32", DenKey: "ansssss", TalkingKey: "qwesad"}
	users := []User{u, u, u, u, u, u, u}
	log.Println(Save(users))
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

func (u User) SetUsername() {
	//TODO implement me
	panic("implement me")
}

func (u User) SetPassword() {
	//TODO implement me
	panic("implement me")
}

type UserHead interface {
	SetUsername()
	SetPassword()
}

//func (u User) setUsername(username string) {
//	u.Username = username
//}
//
//func (u User) setPassword(username string) {
//	u.Password = username
//}
