//go:build !solution

package jsonlist

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
)

func Marshal(w io.Writer, slice any) error {
	if reflect.TypeOf(slice).Kind() != reflect.Slice {
		return &json.UnsupportedTypeError{Type: reflect.TypeOf(slice)}
	}
	v := reflect.ValueOf(slice)
	arrLen := v.Len()
	for i := 0; i < arrLen; i++ {
		elem := v.Index(i).Interface()
		res, err := json.Marshal(elem)
		if err != nil {
			return err
		}
		_, err = w.Write(res)
		if err != nil {
			return err
		}
		if i+1 != arrLen {
			_, err = w.Write([]byte(" "))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func Unmarshal(r io.Reader, slice any) error {
	if reflect.ValueOf(slice).Kind() != reflect.Pointer || reflect.ValueOf(slice).Elem().Kind() != reflect.Slice {
		return &json.UnsupportedTypeError{Type: reflect.TypeOf(slice)}
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	strData := string(data)
	splitted := strings.Split(strData, " ")
	if len(splitted) == 0 {
		slice = nil
		return nil
	}
	cnt := 0
	elemType := reflect.TypeOf(slice).Elem().Elem()
	v := reflect.ValueOf(slice).Elem()
	var cur string
	i := 0
	for _, elem := range splitted {
		cnt += strings.Count(elem, "{")
		cnt -= strings.Count(elem, "}")
		cur += elem
		if cnt == 0 {
			res := reflect.New(elemType).Interface()
			err := json.Unmarshal([]byte(cur), res)
			cur = ""
			if err != nil {
				return err
			}
			v = reflect.Append(v, reflect.ValueOf(res).Elem())
			i++
		}
	}
	reflect.ValueOf(slice).Elem().Set(v)
	return nil
}
