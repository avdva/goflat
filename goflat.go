package goflat

import (
	"reflect"
	"strconv"
)

func flatten(val reflect.Value, prefix string, m map[string]interface{}, visited map[uintptr]struct{}, o *options) {
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
		elem := val.Elem()
		if elem.Kind() == reflect.Pointer && elem.IsNil() {
			m[prefix] = nil
			return
		}
		flatten(val.Elem(), prefix, m, visited, o)
	case reflect.Pointer:
		if val.IsNil() {
			m[prefix] = nil
			return
		}
		var addedPtrs []uintptr
		defer func() {
			for _, ptr := range addedPtrs {
				delete(visited, ptr)
			}
		}()
		elem := val
		for elem.Kind() == reflect.Pointer {
			if _, found := visited[elem.Pointer()]; found {
				return
			}
			visited[elem.Pointer()] = struct{}{}
			addedPtrs = append(addedPtrs, elem.Pointer())
			elem = elem.Elem()
		}
		if val.IsNil() {
			m[prefix] = nil
			return
		}
		switch elem.Kind() {
		case reflect.Struct, reflect.Interface, reflect.Map, reflect.Slice, reflect.Array:
			flatten(elem, prefix, m, visited, o)
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
			tf := typ.Field(i)
			if tf.IsExported() || o.ExpandUnexported {
				flatten(val.Field(i), prefix+typ.Field(i).Name, m, visited, o)
			}
		}
	case reflect.Map:
		if val.IsNil() {
			m[prefix] = nil
			return
		}
		if _, found := visited[val.Pointer()]; found {
			return
		}
		visited[val.Pointer()] = struct{}{}
		defer delete(visited, val.Pointer())
		typ := val.Type()
		if typ.Key().Kind() == reflect.String {
			keys := val.MapKeys()
			if prefix != "" {
				prefix += "."
			}
			for _, key := range keys {
				flatten(val.MapIndex(key), prefix+key.String(), m, visited, o)
			}
		}
	case reflect.Slice, reflect.Array:
		if prefix != "" {
			prefix += "."
		}
		if val.Kind() == reflect.Slice {
			if val.IsNil() {
				m[prefix] = nil
				return
			}
			if _, found := visited[val.Pointer()]; found {
				return
			}
			visited[val.Pointer()] = struct{}{}
			defer delete(visited, val.Pointer())
		}
		for i := 0; i < val.Len(); i++ {
			flatten(val.Index(i), prefix+strconv.Itoa(i), m, visited, o)
		}
	}
}

type options struct {
	ExpandUnexported bool
}

type Option func(o *options)

func ExpandUnexported(expand bool) Option {
	return func(o *options) {
		o.ExpandUnexported = expand
	}
}

// Flatten flattens a golang object.
// It expands structs, maps, slices and arrays, uses '.' as field delimeter.
func Flatten(obj interface{}, opts ...Option) map[string]interface{} {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	val := reflect.ValueOf(obj)
	m := make(map[string]interface{})
	visited := make(map[uintptr]struct{})
	flatten(val, "", m, visited, &o)
	return m
}
