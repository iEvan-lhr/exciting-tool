package tools

import (
	"log"
	"reflect"
	"testing"
	"unsafe"
)

func TestName(t *testing.T) {
	//u := &User{Id: 23132, Username: "foo", Password: "bar", Identity: "324213", QrCodes: nil, DenKey: "ansssss", TalkingKey: "qwesad"}
	////Show(u)
	//s := Save(u)
	//log.Println(s)
	//m := make(map[string]interface{})
	//Lock()
	//for i := 0; i < 10; i++ {
	//	go func(key int) {
	//		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	//		LockFunc("check", func() {
	//			log.Println("Checking:", key)
	//		})
	//	}(i)
	//}
	//time.Sleep(50 * time.Second)
	log.Println(Make("insert userinfo values(").AppendSpiltLR(",", "'", "'", 4, 5, 6, 7, 8, 9, 10).Append(")"))

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
	//s := Make("（林婕琼）")
	//log.Println(Make("12345").FormatterNum())
	//log.Println(s.GetRune("林琼"))
	s := Make("</w:r><w:r w:rsidR=\"008C5277\">\n<w:rPr>\n<w:rFonts w:ascii=\"Century Gothic\" w:eastAsia=\"宋体\" w:hAnsi=\"Century Gothic\" w:cs=\"Times New Roman\" w:hint=\"eastAsia\"/>\n<w:sz w:val=\"18\"/>\n<w:szCs w:val=\"18\"/>\n<w:lang w:eastAsia=\"zh-CN\"/>\n</w:rPr>\n<w:t>(</w:t>\n</w:r>\n\n<w:r w:rsidR=\"008C5277\">\n<w:rPr>\n<w:rFonts w:ascii=\"Century Gothic\" w:eastAsia=\"宋体\" w:hAnsi=\"Century Gothic\" w:cs=\"Times New Roman\"/>\n<w:sz w:val=\"18\"/>\n<w:szCs w:val=\"18\"/>\n<w:lang w:eastAsia=\"zh-CN\"/>\n</w:rPr>\n<w:t>+NMB%)</w:t>\n</w:r>")
	content, other := s.GetContentAll("<w:t>", "</w:t>")
	content[0] = "(+++"
	ans := Make()
	for i := range other {
		ans.appendAny(other[i])
		if i < len(content) {
			ans.appendAny(content[i])
		}
	}
	str := ans.string()
	log.Println(str)
}

func TestTime(t *testing.T) {
	s := Make("09-01-23")
	log.Println(s.UpdateLayout("01-02-06", "2006/01/02"))
	//log.Println(s.GetRune("林琼"))
}
