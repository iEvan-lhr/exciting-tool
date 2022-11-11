package tools

import (
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	start := time.Now()

	m := make(map[*String]int)

	l, l1 := Strings("3212"), Strings("3212")
	log.Println(&l, &l1)
	m[l] = 99
	//Unmarshal("98989898", &s1)
	log.Println(time.Now().Sub(start), m[l1])
	//var list []*String
	//list = append(list, Strings("weq21"))
	//start := time.Now()
	//s1 := "1"
	//for i := 0; i < 99999; i++ {
	//	s1 = s1 + strconv.Itoa(i)
	//}
	//log.Println(time.Now().Sub(start), len(s1))
	//s.AppendAny(19898989)
	//log.Println(s)
}
