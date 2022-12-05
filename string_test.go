package tools

import (
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	log.Println(Make(time.Now()))
}
