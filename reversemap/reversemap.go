//go:build !solution

package reversemap

import (
	"reflect"
)

func ReverseMap(mapa any) any {
	m := reflect.TypeOf(mapa)
	var newMapType = reflect.MapOf(m.Elem(), m.Key())
	newMap := reflect.MakeMapWithSize(newMapType, 0)
	iter := reflect.ValueOf(mapa).MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		newMap.SetMapIndex(v, k)
	}
	return newMap.Interface()
}
