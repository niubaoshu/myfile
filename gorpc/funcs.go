package gorpc

import (
	"reflect"
	"sync"
)

type function struct {
	fType        reflect.Type
	fValue       reflect.Value
	fInNum       int
	fOutNum      int
	fInTypes     []reflect.Type
	fIntypesPool sync.Pool
}

var funcs = []interface{}{}

func Linsten(i interface{}) {
	if reflect.TypeOf(i).Kind() != reflect.Func {
		panic("添加的必须是一个函数")
	}
	funcs = append(funcs, i)
}
