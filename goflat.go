package goflat

import (
	"reflect"
	"strconv"
)

func flatten(val reflect.Value, prefix string, m map[string]interface{}) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String, reflect.Bool:
		if val.CanInterface() {
			if prefix == "" {
				prefix = val.Type().Name()
			}
			m[prefix] = val.Interface()
		}
	case reflect.Interface:
		flatten(val.Elem(), prefix, m)
	case reflect.Pointer:
		elem := val.Elem()
		for elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		switch elem.Kind() {
		case reflect.Struct, reflect.Interface, reflect.Map, reflect.Slice, reflect.Array:
			flatten(elem, prefix, m)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
			reflect.String, reflect.Bool:
			if val.CanInterface() {
				m[prefix] = val.Interface()
			}
		}
	case reflect.Struct:
		typ := val.Type()
		if prefix != "" {
			prefix += "."
		}
		for i := 0; i < typ.NumField(); i++ {
			flatten(val.Field(i), prefix+typ.Field(i).Name, m)
		}
	case reflect.Map:
		typ := val.Type()
		if typ.Key().Kind() == reflect.String {
			keys := val.MapKeys()
			if prefix != "" {
				prefix += "."
			}
			for _, key := range keys {
				flatten(val.MapIndex(key), prefix+key.String(), m)
			}
		}
	case reflect.Slice, reflect.Array:
		if prefix != "" {
			prefix += "."
		}
		for i := 0; i < val.Len(); i++ {
			flatten(val.Index(i), prefix+strconv.Itoa(i), m)
		}
	}
}

// Flatten flattens a golang object.
// It expands structs, maps, slices and arrays, uses '.' as field delimeter.
func Flatten(obj interface{}) map[string]interface{} {
	val := reflect.ValueOf(obj)
	m := make(map[string]interface{})
	flatten(val, "", m)
	return m
}
