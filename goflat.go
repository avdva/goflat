package goflat

import (
	"reflect"
	"sort"
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

func makeOptions(opts ...Option) *options {
	options := &options{
		delimeter: ".",
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func (w *walker) run(val reflect.Value) {
	path := make([]string, 0, 16)
	w.visit(val, path)
}

func (w *walker) visit(val reflect.Value, path []string) {
	switch kind := val.Kind(); {
	case kind >= reflect.Int && kind <= reflect.Int64:
		w.visitInt(val, path)
	case kind >= reflect.Uint8 && kind <= reflect.Uint64:
		w.visitUint(val, path)
	case kind == reflect.Float32:
		w.visitPrimitive(float32(val.Float()), path)
	case kind == reflect.Float64:
		w.visitPrimitive(val.Float(), path)
	case kind == reflect.Bool:
		w.visitPrimitive(val.Bool(), path)
	case kind == reflect.Complex64:
		w.visitPrimitive(complex64(val.Complex()), path)
	case kind == reflect.Complex128:
		w.visitPrimitive(val.Complex(), path)
	case kind == reflect.String:
		w.visitPrimitive(val.String(), path)
	case kind == reflect.Interface:
		w.visit(val.Elem(), path)
	case kind == reflect.Pointer:
		w.visitPointer(val, path)
	case kind == reflect.Struct:
		w.visitStruct(val, path)
	case kind == reflect.Map:
		w.visitMap(val, path)
	case kind == reflect.Slice || kind == reflect.Array:
		w.visitSliceOrArray(val, path)
	}
}

func (w *walker) visitInt(val reflect.Value, path []string) {
	iVal := val.Int()
	switch val.Kind() {
	case reflect.Int:
		w.visitPrimitive(int(iVal), path)
	case reflect.Int8:
		w.visitPrimitive(int8(iVal), path)
	case reflect.Int16:
		w.visitPrimitive(int16(iVal), path)
	case reflect.Int32:
		w.visitPrimitive(int32(iVal), path)
	case reflect.Int64:
		w.visitPrimitive(int64(iVal), path)
	}
}

func (w *walker) visitUint(val reflect.Value, path []string) {
	iVal := val.Uint()
	switch val.Kind() {
	case reflect.Uint:
		w.visitPrimitive(uint(iVal), path)
	case reflect.Uint8:
		w.visitPrimitive(uint8(iVal), path)
	case reflect.Uint16:
		w.visitPrimitive(uint16(iVal), path)
	case reflect.Uint32:
		w.visitPrimitive(uint32(iVal), path)
	case reflect.Uint64:
		w.visitPrimitive(uint64(iVal), path)
	}
}

func (w *walker) visitPrimitive(val interface{}, path []string) {
	if len(path) == 0 {
		path = []string{"."}
	}
	w.cb(path, val)
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
		if w.o.sortMapKeys {
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].String() < keys[j].String()
			})
		}
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
	sortMapKeys      bool
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

// WithDelimeter option sets a field delimeter. '.' is the default delimeter.
func WithDelimeter(delim string) Option {
	return func(o *options) {
		o.delimeter = delim
	}
}

// SortMapKeys option, if set, will force sorting of map keys before visiting.
func SortMapKeys(sort bool) Option {
	return func(o *options) {
		o.sortMapKeys = sort
	}
}

// Flatten flattens a golang object.
// It expands structs, maps, slices and arrays, uses '.' as a default field delimeter.
func Flatten(obj interface{}, opts ...Option) map[string]interface{} {
	m := make(map[string]interface{})
	o := makeOptions(opts...)
	w := newWalker(func(path []string, value interface{}) bool {
		m[strings.Join(path, o.delimeter)] = value
		return true
	}, o)
	w.run(reflect.ValueOf(obj))
	return m
}

// WalkFunc is a callback to be called for each value.
type WalkFunc func(path []string, value interface{}) bool

// Walk calls cb for every member field of the obj.
func Walk(obj interface{}, cb WalkFunc, opts ...Option) {
	w := newWalker(cb, makeOptions(opts...))
	w.run(reflect.ValueOf(obj))
}
