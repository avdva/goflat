package goflat

import (
	"reflect"
	"strconv"
)

type walker struct {
	m       map[string]interface{}
	visited map[uintptr]struct{}
	o       *options
}

func newWalker() *walker {
	return &walker{
		m:       make(map[string]interface{}),
		visited: make(map[uintptr]struct{}),
		o: &options{
			delimeter:        ".",
			expandUnexported: false,
		},
	}
}

func (w *walker) visit(val reflect.Value, prefix string) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String, reflect.Bool:
		w.visitSimple(val, prefix)
	case reflect.Interface:
		w.visit(val.Elem(), prefix)
	case reflect.Pointer:
		w.visitPointer(val, prefix)
	case reflect.Struct:
		w.visitStruct(val, prefix)
	case reflect.Map:
		w.visitMap(val, prefix)
	case reflect.Slice, reflect.Array:
		w.visitSliceOrArray(val, prefix)
	}
}

func (w *walker) visitSimple(val reflect.Value, prefix string) {
	if val.CanInterface() {
		if prefix == "" {
			prefix = val.Type().Name()
		}
		w.m[prefix] = val.Interface()
	}
}

func (w *walker) visitPointer(val reflect.Value, prefix string) {
	var addedPtrs []uintptr
	defer func() {
		for _, ptr := range addedPtrs {
			delete(w.visited, ptr)
		}
	}()
	elem := val
	var isNil bool
	for elem.Kind() == reflect.Pointer {
		if _, found := w.visited[elem.Pointer()]; found {
			return
		}
		if elem.IsNil() {
			isNil = true
		}
		w.visited[elem.Pointer()] = struct{}{}
		addedPtrs = append(addedPtrs, elem.Pointer())
		elem = elem.Elem()
	}
	switch elem.Kind() {
	case reflect.Struct, reflect.Interface, reflect.Map, reflect.Slice, reflect.Array, reflect.Invalid:
		if !isNil {
			w.visit(elem, prefix)
		} else if w.o.addNilContainers {
			w.m[prefix] = nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.String, reflect.Bool:
		if !isNil {
			if val.CanInterface() {
				w.m[prefix] = val.Interface()
			}
		} else if w.o.addNilFields {
			w.m[prefix] = nil
		}
	}
}

func (w *walker) visitStruct(val reflect.Value, prefix string) {
	typ := val.Type()
	if prefix != "" {
		prefix += w.o.delimeter
	}
	for i := 0; i < typ.NumField(); i++ {
		tf := typ.Field(i)
		if tf.IsExported() || w.o.expandUnexported {
			w.visit(val.Field(i), prefix+typ.Field(i).Name)
		}
	}
}

func (w *walker) visitMap(val reflect.Value, prefix string) {
	if val.IsNil() {
		if w.o.addNilContainers {
			w.m[prefix] = nil
		}
		return
	}
	if _, found := w.visited[val.Pointer()]; found {
		return
	}
	w.visited[val.Pointer()] = struct{}{}
	defer delete(w.visited, val.Pointer())
	typ := val.Type()
	if typ.Key().Kind() == reflect.String {
		keys := val.MapKeys()
		if prefix != "" {
			prefix += w.o.delimeter
		}
		for _, key := range keys {
			w.visit(val.MapIndex(key), prefix+key.String())
		}
	}
}

func (w *walker) visitSliceOrArray(val reflect.Value, prefix string) {
	if val.Kind() == reflect.Slice {
		if val.IsNil() {
			if w.o.addNilContainers {
				w.m[prefix] = nil
			}
			return
		}
		if _, found := w.visited[val.Pointer()]; found {
			return
		}
		w.visited[val.Pointer()] = struct{}{}
		defer delete(w.visited, val.Pointer())
	}
	if prefix != "" {
		prefix += w.o.delimeter
	}
	for i := 0; i < val.Len(); i++ {
		w.visit(val.Index(i), prefix+strconv.Itoa(i))
	}
}

type options struct {
	expandUnexported bool
	delimeter        string
	addNilContainers bool
	addNilFields     bool
}

// Option allows to customise flattening
type Option func(o *options)

// ExpandUnexported option controls how to deal with non exported struct fields.
// By default, such fields are not expanded.
func ExpandUnexported(expand bool) Option {
	return func(o *options) {
		o.expandUnexported = expand
	}
}

// AddNilContainers option, if set, allows to add en entry for nil maps, slices, arrays, interfaces.
func AddNilContainers(add bool) Option {
	return func(o *options) {
		o.addNilContainers = add
	}
}

// AddNilContainers option, if set, allows to add en entry for nil ints, floats, strings.
func AddNilFields(add bool) Option {
	return func(o *options) {
		o.addNilFields = add
	}
}

// WithDelimeter option set a field delimeter. '.' is the default delimeter.
func WithDelimeter(delim string) Option {
	return func(o *options) {
		o.delimeter = delim
	}
}

// Flatten flattens a golang object.
// It expands structs, maps, slices and arrays, uses '.' as a default field delimeter.
func Flatten(obj interface{}, opts ...Option) map[string]interface{} {
	w := newWalker()
	for _, opt := range opts {
		opt(w.o)
	}
	val := reflect.ValueOf(obj)
	w.visit(val, "")
	return w.m
}
