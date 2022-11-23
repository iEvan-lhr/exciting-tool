package tools

import (
	"log"
	"reflect"
)

type ParseError struct {
	values []reflect.Value
	err    error
	isErr  bool
}

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

func ReturnError(err error, args ...interface{}) (vars *ParseError) {
	vars = &ParseError{}
	if err != nil {
		if len(args) <= 2 {
			panic(err)
		}
		vars.isErr = true
		var values []reflect.Value
		for _, arg := range args[3].([]interface{}) {
			values = append(values, reflect.ValueOf(arg))
		}
		vars.values = append(vars.values, reflect.ValueOf(args[2]).Call(values)...)
	} else {
		vars.isErr = false
		var values []reflect.Value
		for _, arg := range args[1].([]interface{}) {
			values = append(values, reflect.ValueOf(arg))
		}
		vars.values = append(vars.values, reflect.ValueOf(args[0]).Call(values)...)
	}
	return
}

func (p *ParseError) Unmarshal(args ...interface{}) {
	switch args[0].(type) {
	case []interface{}:
		if p.isErr {
			if len(args[1].([]interface{})) == 0 {
				panic("ParseErrorFail:error message:" + p.err.Error())
			}
			for i, v := range args[1].([]interface{}) {
				reflect.ValueOf(v).Elem().Set(p.values[i])
			}
		} else {
			for i, v := range args[0].([]interface{}) {
				reflect.ValueOf(v).Elem().Set(p.values[i])
			}
		}
	default:
		for i, v := range args {
			reflect.ValueOf(v).Elem().Set(p.values[i])
		}
	}
}
