package tools

import (
	"log"
	"strconv"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	var sli []*String
	for i := 0; i < 99999; i++ {
		s := &String{}
		s.AppendAny(i)
		sli = append(sli, s)
	}
	start := time.Now()
	for _, v := range sli {
		_, err := v.Atoi()
		if err != nil {
			panic(err)
		}
	}
	log.Println(time.Now().Sub(start))
}

func TestAtoi(t *testing.T) {
	var sli []string
	for i := 0; i < 99999; i++ {
		sli = append(sli, strconv.Itoa(i))
	}
	start := time.Now()
	for _, v := range sli {
		_, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
	}
	log.Println(time.Now().Sub(start))
}
