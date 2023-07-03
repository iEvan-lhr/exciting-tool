package tools

import (
	"log"
	"time"
)

type LockA struct {
	unix time.Time
	Name string `json:"name"`
}

var lockMission map[string]chan struct{}
var lockFunc map[string]chan func()

func Lock(l any) LockA {
	switch l.(type) {
	case string:
		do(l.(string))
		return LockA{unix: time.Now(), Name: l.(string)}
	case func():
		log.Println("Lock Func")
	case func(lock LockA):
		log.Println("Lock Func")
	case struct{}:
		log.Println("Lock Struct")
	}
	return LockA{}
}

func do(name string) {
	if lockMission == nil || len(lockMission) == 0 {
		lockMission = make(map[string]chan struct{})
	}
	if _, ok := lockMission[name]; !ok {
		lockMission[name] = make(chan struct{}, 3)
		go func() {
			for {
				_, ok := <-lockMission[name]
				if !ok {
					delete(lockMission, name)
				} else {
					log.Println("check")
				}
			}
		}()
	}
	lockMission[name] <- struct{}{}
}

func LockFunc(name string, f func()) {
	if lockFunc == nil || len(lockFunc) == 0 {
		lockFunc = make(map[string]chan func())
	}
	if _, ok := lockFunc[name]; !ok {
		lockFunc[name] = make(chan func(), 3)
		go func() {
			for {
				v, ok := <-lockFunc[name]
				if !ok {
					delete(lockFunc, name)
				} else {
					v()
				}
			}
		}()
	}
	lockFunc[name] <- f
}
