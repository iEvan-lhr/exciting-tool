package tools

import (
	"log"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	//spider := MakeSpider(0, "999999")
	m := make(map[string]int)
	values := 999998
	for i := 1; i < 999999; i++ {
		//spider.Add(i, 999998)
		m[strconv.Itoa(i)] = values
		values--
	}
	//log.Println(spider.len)

}

func TestSpider(t *testing.T) {
	spider := MakeSpider(0, "999999")
	//m := make(map[string]int)
	values := 999998
	for i := 1; i < 999999; i++ {
		spider.Add(i, values)
		//m[strconv.Itoa(i)] = values
		values--
	}
	log.Println(spider.Get(232))

}
