package tools

import (
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	start := time.Now()
	s := Strings("1")
	for i := 0; i < 99999; i++ {
		s.AppendAny(i)
	}
	log.Println(time.Now().Sub(start), s.Len())
	//start := time.Now()
	//s1 := "1"
	//for i := 0; i < 99999; i++ {
	//	s1 = s1 + strconv.Itoa(i)
	//}
	//log.Println(time.Now().Sub(start), len(s1))
	//s.AppendAny(19898989)
	//log.Println(s)
}
