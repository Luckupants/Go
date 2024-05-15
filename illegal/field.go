//go:build !solution

package illegal

import (
	"reflect"
	"unsafe"
)

func SetPrivateField(obj interface{}, name string, value interface{}) {
	v := reflect.ValueOf(obj)
	f, has := reflect.TypeOf(obj).Elem().FieldByName(name)
	if !has || f.Type != reflect.TypeOf(value) {
		panic("ti daun??")
	}
	p := unsafe.Pointer((uintptr)(v.UnsafePointer()) + f.Offset)
	vv := reflect.NewAt(reflect.TypeOf(value), p)
	vv.Elem().Set(reflect.ValueOf(value))
}
