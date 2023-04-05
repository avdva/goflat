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

func (w *walker) run(val reflect.Value) {
	path := make([]string, 0, 16)
	w.visit(val, path)
}

func (w *walker) visit(val reflect.Value, path []string) (cont bool) {
	switch kind := val.Kind(); {
	case kind >= reflect.Int && kind <= reflect.Int64:
		cont = w.visitInt(val, path)
	case kind >= reflect.Uint8 && kind <= reflect.Uint64:
		cont = w.visitUint(val, path)
	case kind == reflect.Float32:
		cont = w.visitPrimitive(float32(val.Float()), path)
	case kind == reflect.Float64:
		cont = w.visitPrimitive(val.Float(), path)
	case kind == reflect.Bool:
		cont = w.visitPrimitive(val.Bool(), path)
	case kind == reflect.Complex64:
		cont = w.visitPrimitive(complex64(val.Complex()), path)
	case kind == reflect.Complex128:
		cont = w.visitPrimitive(val.Complex(), path)
	case kind == reflect.String:
		cont = w.visitPrimitive(val.String(), path)
	case kind == reflect.Interface:
		cont = w.visit(val.Elem(), path)
	case kind == reflect.Pointer:
		cont = w.visitPointer(val, path)
	case kind == reflect.Struct:
		cont = w.visitStruct(val, path)
	case kind == reflect.Map:
		cont = w.visitMap(val, path)
	case kind == reflect.Slice || kind == reflect.Array:
		cont = w.visitSliceOrArray(val, path)
	}
	return cont
}

func (w *walker) visitInt(val reflect.Value, path []string) (cont bool) {
	iVal := val.Int()
	switch val.Kind() {
	case reflect.Int:
		cont = w.visitPrimitive(int(iVal), path)
	case reflect.Int8:
		cont = w.visitPrimitive(int8(iVal), path)
	case reflect.Int16:
		cont = w.visitPrimitive(int16(iVal), path)
	case reflect.Int32:
		cont = w.visitPrimitive(int32(iVal), path)
	case reflect.Int64:
		cont = w.visitPrimitive(int64(iVal), path)
	}
	return cont
}

func (w *walker) visitUint(val reflect.Value, path []string) (cont bool) {
	iVal := val.Uint()
	switch val.Kind() {
	case reflect.Uint:
		cont = w.visitPrimitive(uint(iVal), path)
	case reflect.Uint8:
		cont = w.visitPrimitive(uint8(iVal), path)
	case reflect.Uint16:
		cont = w.visitPrimitive(uint16(iVal), path)
	case reflect.Uint32:
		cont = w.visitPrimitive(uint32(iVal), path)
	case reflect.Uint64:
		cont = w.visitPrimitive(uint64(iVal), path)
	}
	return cont
}

func (w *walker) visitPrimitive(val interface{}, path []string) (cont bool) {
	return w.cb(path, val)
}

func (w *walker) visitPointer(val reflect.Value, path []string) (cont bool) {
	var addedPtrs []uintptr
	defer func() {
		for _, ptr := range addedPtrs {
			delete(w.visited, ptr)
		}
	}()
	var isNil bool
	elem := val
	for ; elem.Kind() == reflect.Pointer; elem = elem.Elem() {
		if _, found := w.visited[elem.Pointer()]; found {
			return true
		}
		if elem.IsNil() {
			isNil = true
		}
		w.visited[elem.Pointer()] = struct{}{}
		addedPtrs = append(addedPtrs, elem.Pointer())
	}
	if isNil {
		if val.CanInterface() {
			return w.cb(path, val.Interface())
		}
		return w.cb(path, reflect.New(val.Type()).Elem().Interface())
	}
	switch w.o.pointerFollowPolicy {
	case PointerPolicyJustPointer:
		if val.CanInterface() {
			if !w.cb(path, val.Interface()) {
				return
			}
		}
	case PointerPolicyPrimitivePointer:
		if isPrimitive(elem.Kind()) {
			if val.CanInterface() {
				if !w.cb(path, val.Interface()) {
					return
				}
			}
		} else {
			return w.visit(elem, path)
		}
	case PointerPolicyBoth:
		if val.CanInterface() {
			if !w.cb(path, val.Interface()) {
				return
			}
		}
		fallthrough
	case PointerPolicyJustValue:
		return w.visit(elem, path)
	}
	return true
}

func isPrimitive(kind reflect.Kind) bool {
	primitives := []bool{
		true,  // Invalid
		true,  // Bool
		true,  // Int
		true,  // Int8
		true,  // Int16
		true,  // Int32
		true,  // Int64
		true,  // Uint
		true,  // Uint8
		true,  // Uint16
		true,  // Uint32
		true,  // Uint64
		true,  // Uintptr
		true,  // Float32
		true,  // Float64
		true,  // Complex64
		true,  // Complex128
		false, // Array
		false, // Chan
		false, // Func
		false, // Interface
		false, // Map
		true,  // Pointer
		false, // Slice
		true,  // String
		false, // Struct
		true,  // UnsafePointer
	}
	return primitives[int(kind)]
}

func (w *walker) visitStruct(val reflect.Value, path []string) (cont bool) {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		tf := typ.Field(i)
		if tf.IsExported() || w.o.expandUnexported {
			if !w.visit(val.Field(i), append(path, typ.Field(i).Name)) {
				return false
			}
		}
	}
	return true
}

func (w *walker) visitMap(val reflect.Value, path []string) (cont bool) {
	if val.IsNil() {
		if w.o.addNilContainers {
			return w.cb(path, nil)
		}
		return true
	}
	if _, found := w.visited[val.Pointer()]; found {
		return true
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
			if !w.visit(val.MapIndex(key), append(path, key.String())) {
				return false
			}
		}
	}
	return true
}

func (w *walker) visitSliceOrArray(val reflect.Value, path []string) (cont bool) {
	if val.Kind() == reflect.Slice {
		if val.IsNil() {
			if w.o.addNilContainers {
				return w.cb(path, nil)
			}
			return true
		}
		if _, found := w.visited[val.Pointer()]; found {
			return true
		}
		w.visited[val.Pointer()] = struct{}{}
		defer delete(w.visited, val.Pointer())
	}
	for i := 0; i < val.Len(); i++ {
		if !w.visit(val.Index(i), append(path, strconv.Itoa(i))) {
			return false
		}
	}
	return true
}

const (
	// PointerPolicyBoth: WalkFunc will be called for both pointer (when possible) and underlying value.
	PointerPolicyBoth = iota
	// PointerPolicyJustPointer: WalkFunc will be called for pointer only (when possible).
	// To walk the underlying value, call Walk() with that value should be called.
	PointerPolicyJustPointer
	// PointerPolicyJustValue: WalkFunc will be called for the value only.
	PointerPolicyJustValue
	// PointerPolicyPrimitivePointer: WalkFunc will be called with a pointer arg only for
	// simple data types like ints, floats, complex's, booleans, pointers, strings.
	// For more complex types (e.g. structs) Walk follows the pointer and calls WalkFunc
	// with underlying objects only. This is the default policy.
	PointerPolicyPrimitivePointer
)

type options struct {
	expandUnexported    bool
	delimeter           string
	addNilContainers    bool
	addNilFields        bool
	sortMapKeys         bool
	pointerFollowPolicy int8
}

func makeOptions(opts ...Option) *options {
	options := &options{
		delimeter:           ".",
		pointerFollowPolicy: PointerPolicyPrimitivePointer,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
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

// WithPointerFllowPolicy option specifies how to handle pointer types. See PointerPolicy* consts.
func WithPointerFllowPolicy(policy int8) Option {
	return func(o *options) {
		o.pointerFollowPolicy = policy
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
