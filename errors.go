package errors

import (
	"log"
	"reflect"
)

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
}

func ExecError(err error) {
	if err != nil {
		panic(err)
	}
}

func LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}
