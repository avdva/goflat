package goflat

import (
	"reflect"
	"strconv"
	"strings"
)

type walker struct {
	cb      WalkFunc
	visited map[uintptr]struct{}
	o       *options
}

func newWalker(cb WalkFunc, o *options) *walker {
	return &walker{
		cb:      cb,
		visited: make(map[uintptr]struct{}),
		o:       o,
	}
}

func defaultOptions() *options {
	return &options{
		delimeter:        ".",
		expandUnexported: false,
	}
}

func (w *walker) run(val reflect.Value) {
	path := make([]string, 0, 16)
	w.visit(val, path)
}

func (w *walker) visit(val reflect.Value, path []string) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String, reflect.Bool:
		w.visitPrimitive(val, path)
	case reflect.Interface:
		w.visit(val.Elem(), path)
	case reflect.Pointer:
		w.visitPointer(val, path)
	case reflect.Struct:
		w.visitStruct(val, path)
	case reflect.Map:
		w.visitMap(val, path)
	case reflect.Slice, reflect.Array:
		w.visitSliceOrArray(val, path)
	}
}

func (w *walker) visitPrimitive(val reflect.Value, path []string) {
	if val.CanInterface() {
		if len(path) == 0 {
			path = append(path, val.Type().Name())
		}
		w.cb(path, val.Interface())
	}
}

func (w *walker) visitPointer(val reflect.Value, path []string) {
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
			w.visit(elem, path)
		} else if w.o.addNilContainers {
			w.cb(path, nil)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.String, reflect.Bool:
		if !isNil {
			if val.CanInterface() {
				w.cb(path, val.Interface())
			}
		} else if w.o.addNilFields {
			w.cb(path, nil)
		}
	}
}

func (w *walker) visitStruct(val reflect.Value, path []string) {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		tf := typ.Field(i)
		if tf.IsExported() || w.o.expandUnexported {
			w.visit(val.Field(i), append(path, typ.Field(i).Name))
		}
	}
}

func (w *walker) visitMap(val reflect.Value, path []string) {
	if val.IsNil() {
		if w.o.addNilContainers {
			w.cb(path, nil)
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
		for _, key := range keys {
			w.visit(val.MapIndex(key), append(path, key.String()))
		}
	}
}

func (w *walker) visitSliceOrArray(val reflect.Value, path []string) {
	if val.Kind() == reflect.Slice {
		if val.IsNil() {
			if w.o.addNilContainers {
				w.cb(path, nil)
			}
			return
		}
		if _, found := w.visited[val.Pointer()]; found {
			return
		}
		w.visited[val.Pointer()] = struct{}{}
		defer delete(w.visited, val.Pointer())
	}
	for i := 0; i < val.Len(); i++ {
		w.visit(val.Index(i), append(path, strconv.Itoa(i)))
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
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	m := make(map[string]interface{})
	w := newWalker(func(path []string, value interface{}) {
		m[strings.Join(path, o.delimeter)] = value
	}, o)
	val := reflect.ValueOf(obj)
	w.run(val)
	return m
}

// WalkFunc is a callback to be called for each value.
type WalkFunc func(path []string, value interface{})

// Walk calls cb for every member field of the obj.
func Walk(obj interface{}, cb WalkFunc, opts ...Option) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	w := newWalker(cb, o)
	val := reflect.ValueOf(obj)
	w.run(val)
}
