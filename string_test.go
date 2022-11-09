package tools

import (
	"log"
	"testing"
)

func TestName(t *testing.T) {
	s := Strings("asd")
	log.Println(s.Len(), s.Check("asd"), s.Check("Asd"))
	s.FirstUpper()
	log.Println(s, s.Check(Strings("Asd")))
	s.RemoveIndexStr(1)
	s.Append("sajdajewjwahejwae")
	log.Println(s)
}
