package tools

import (
	"log"
	"reflect"
)

func ReturnValue(v ...interface{}) interface{} {
	if v[len(v)-1] != nil {
		log.Println(v[len(v)-1])
	}
	return v[0]
}

func ReturnValueByTwo(v ...interface{}) interface{} {
	if v[len(v)-1] != nil {
		panic(v[len(v)-1])
	}
	return v[0]
}

func PanicError(f ...func() error) {
	for _, fun := range f {
		if err := fun(); err != nil {
			panic(err)
		}
	}
}

// ExecGoFunc       方法                参数
func ExecGoFunc(exec interface{}, args ...interface{}) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				panic(e)
			}
		}()
		var values []reflect.Value
		for _, arg := range args {
			values = append(values, reflect.ValueOf(arg))
		}
		reflect.ValueOf(exec).Call(values)
	}()
}

func ExecError(err error) {
	if err != nil {
		panic(err)
	}
}

func Error(e interface{}) {
	switch e.(type) {
	case error:
		if e.(error) != nil {
			panic(e.(error))
		}
	case []error:
		for _, err := range e.([]error) {
			if err != nil {
				panic(err)
			}
		}
	}
}

func LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func DeferError(err error, exec interface{}, args ...interface{}) {
	defer func() {
		var values []reflect.Value
		for _, arg := range args {
			values = append(values, reflect.ValueOf(arg))
		}
		reflect.ValueOf(exec).Call(values)
	}()
	if err != nil {
		panic(err)
	}
}
