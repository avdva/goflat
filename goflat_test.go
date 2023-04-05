package goflat

import (
	"strconv"
	"strings"
	"testing"

	"github.com/avdva/goflat/testpkg"

	"github.com/stretchr/testify/assert"
)

type pathValue struct {
	path  []string
	value interface{}
}

func TestPointerPolicy(t *testing.T) {
	ts := testpkg.NewPointerTestStruct()

	tests := []struct {
		obj    interface{}
		exp    []pathValue
		policy int8
	}{
		{
			obj:    ts,
			policy: PointerPolicyJustPointer,
			exp: []pathValue{
				{path: []string{}, value: ts},
			},
		},
		{
			obj:    *ts,
			policy: PointerPolicyJustPointer,
			exp: []pathValue{
				{path: []string{"IntPtr"}, value: ts.IntPtr},
				{path: []string{"IntPtrPtr"}, value: ts.IntPtrPtr},
				{path: []string{"unexportedNil"}, value: (*string)(nil)},
				{path: []string{"ExportedStruct"}, value: ts.ExportedStruct},
				{path: []string{"unexportedNilStruct"}, value: (*testpkg.SmallStruct)(nil)},
				{path: []string{"unexportedSlice", "0"}, value: 1},
				{path: []string{"unexportedSlice", "1"}, value: 2},
				{path: []string{"unexportedSlice", "2"}, value: 3},
			},
		},
		{
			obj:    ts,
			policy: PointerPolicyJustValue,
			exp: []pathValue{
				{path: []string{"IntPtr"}, value: 5},
				{path: []string{"IntPtrPtr"}, value: 5},
				{path: []string{"unexportedString"}, value: "hello"},
				{path: []string{"unexportedNil"}, value: (*string)(nil)},
				{path: []string{"ExportedStruct", "Data"}, value: 123.456},
				{path: []string{"unexportedStruct", "Data"}, value: 123.456},
				{path: []string{"unexportedNilStruct"}, value: (*testpkg.SmallStruct)(nil)},
				{path: []string{"unexportedSlice", "0"}, value: 1},
				{path: []string{"unexportedSlice", "1"}, value: 2},
				{path: []string{"unexportedSlice", "2"}, value: 3},
			},
		},
		{
			obj:    ts,
			policy: PointerPolicyPrimitivePointer,
			exp: []pathValue{
				{path: []string{"IntPtr"}, value: ts.IntPtr},
				{path: []string{"IntPtrPtr"}, value: ts.IntPtrPtr},
				{path: []string{"unexportedNil"}, value: (*string)(nil)},
				{path: []string{"ExportedStruct", "Data"}, value: ts.ExportedStruct.Data},
				{path: []string{"unexportedNilStruct"}, value: (*testpkg.SmallStruct)(nil)},
				{path: []string{"unexportedSlice", "0"}, value: 1},
				{path: []string{"unexportedSlice", "1"}, value: 2},
				{path: []string{"unexportedSlice", "2"}, value: 3},
			},
		},
		{
			obj:    ts,
			policy: PointerPolicyBoth,
			exp: []pathValue{
				{path: []string{}, value: ts},
				{path: []string{"IntPtr"}, value: ts.IntPtr},
				{path: []string{"IntPtr"}, value: 5},
				{path: []string{"IntPtrPtr"}, value: ts.IntPtrPtr},
				{path: []string{"IntPtrPtr"}, value: 5},
				{path: []string{"unexportedString"}, value: "hello"},
				{path: []string{"unexportedNil"}, value: (*string)(nil)},
				{path: []string{"ExportedStruct"}, value: ts.ExportedStruct},
				{path: []string{"ExportedStruct", "Data"}, value: ts.ExportedStruct.Data},
				{path: []string{"ExportedStruct", "Data"}, value: 123.456},
				{path: []string{"unexportedStruct", "Data"}, value: 123.456},
				{path: []string{"unexportedNilStruct"}, value: (*testpkg.SmallStruct)(nil)},
				{path: []string{"unexportedSlice", "0"}, value: 1},
				{path: []string{"unexportedSlice", "1"}, value: 2},
				{path: []string{"unexportedSlice", "2"}, value: 3},
			},
		},
	}
	commonOpts := []Option{
		ExpandUnexported(true),
		AddNilContainers(true),
		AddNilFields(true),
		SortMapKeys(true),
	}
	for i := range tests {
		idx := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			test := tests[idx]
			a := assert.New(t)
			var slice []pathValue
			Walk(test.obj, func(path []string, value interface{}) bool {
				p := make([]string, len(path))
				copy(p, path)
				slice = append(slice, pathValue{path: p, value: value})
				return true
			}, append(commonOpts, WithPointerFllowPolicy(test.policy))...)
			a.Equal(test.exp, slice)
		})
	}
}

func TestWalkFlatten(t *testing.T) {
	ts := testpkg.NewTestStruct()

	tests := []struct {
		obj  interface{}
		exp  []pathValue
		opts []Option
	}{
		{
			obj: ts,
			exp: []pathValue{
				{
					path:  []string{},
					value: ts,
				},
				{
					path:  []string{"A"},
					value: 5,
				},
				{
					path:  []string{"B"},
					value: uint64(6),
				},
				{
					path:  []string{"S", "D"},
					value: "D",
				},
				{
					path:  []string{"S", "Ptr"},
					value: ts.S.Ptr,
				},
				{
					path:  []string{"S", "Ptr"},
					value: *ts.S.Ptr,
				},
				{
					path:  []string{"S", "M", "k"},
					value: 123,
				},
				{
					path:  []string{"S", "notExportedMap", "k"},
					value: "v",
				},
				{
					path:  []string{"Nested", "Val"},
					value: true,
				},
				{
					path:  []string{"embedded", "S"},
					value: int8(123),
				},
				{
					path:  []string{"Iface"},
					value: ts.Iface,
				},
				{
					path:  []string{"Iface", "Val"},
					value: "iface",
				},
				{
					path:  []string{"PtrPtr"},
					value: ts.PtrPtr,
				},
				{
					path:  []string{"PtrPtr"},
					value: **ts.PtrPtr,
				},
				{
					path:  []string{"M"},
					value: ts.M,
				},
				{
					path:  []string{"M", "key"},
					value: "value",
				},
				{
					path:  []string{"Slice", "0"},
					value: float64(26.05),
				},
				{
					path:  []string{"Slice", "1"},
					value: float64(1.1),
				},
				{
					path:  []string{"Slice", "2"},
					value: float64(23.12),
				},
				{
					path:  []string{"Array", "0"},
					value: float32(1),
				},
				{
					path:  []string{"Array", "1"},
					value: float32(2),
				},
				{
					path:  []string{"Array", "2"},
					value: float32(3),
				},
				{
					path:  []string{"NilSlice"},
					value: nil,
				},
				{
					path:  []string{"Complex64"},
					value: complex64(complex(1, 2)),
				},
				{
					path:  []string{"Complex128"},
					value: complex(3, 4),
				},
				{
					path:  []string{"notExportedInt"},
					value: 123,
				},
				{
					path:  []string{"notExportedIface", "Val"},
					value: "string",
				},
				{
					path:  []string{"notExportedStruct", "A"},
					value: 0,
				},
				{
					path:  []string{"notExportedPointer"},
					value: "string",
				},
			},
			opts: []Option{
				ExpandUnexported(true),
				AddNilContainers(true),
				AddNilFields(true),
				SortMapKeys(true),
				WithPointerFllowPolicy(PointerPolicyBoth),
			},
		},
		{
			obj: ts,
			exp: []pathValue{
				{
					path:  []string{"A"},
					value: 5,
				},
				{
					path:  []string{"B"},
					value: uint64(6),
				},
				{
					path:  []string{"S", "D"},
					value: "D",
				},
				{
					path:  []string{"S", "Ptr"},
					value: ts.S.Ptr,
				},
				{
					path:  []string{"S", "M", "k"},
					value: 123,
				},
				{
					path:  []string{"Nested", "Val"},
					value: true,
				},
				{
					path:  []string{"Iface", "Val"},
					value: "iface",
				},
				{
					path:  []string{"PtrPtr"},
					value: ts.PtrPtr,
				},
				{
					path:  []string{"M", "key"},
					value: "value",
				},
				{
					path:  []string{"Slice", "0"},
					value: float64(26.05),
				},
				{
					path:  []string{"Slice", "1"},
					value: float64(1.1),
				},
				{
					path:  []string{"Slice", "2"},
					value: float64(23.12),
				},
				{
					path:  []string{"Array", "0"},
					value: float32(1),
				},
				{
					path:  []string{"Array", "1"},
					value: float32(2),
				},
				{
					path:  []string{"Array", "2"},
					value: float32(3),
				},
				{
					path:  []string{"NilSlice"},
					value: nil,
				},
				{
					path:  []string{"Complex64"},
					value: complex64(complex(1, 2)),
				},
				{
					path:  []string{"Complex128"},
					value: complex(3, 4),
				},
			},
			opts: []Option{
				ExpandUnexported(false),
				AddNilContainers(true),
				AddNilFields(false),
				WithPointerFllowPolicy(PointerPolicyPrimitivePointer),
			},
		},
		{
			obj: ts,
			exp: []pathValue{
				{
					path:  []string{"A"},
					value: 5,
				},
				{
					path:  []string{"B"},
					value: uint64(6),
				},
				{
					path:  []string{"S", "D"},
					value: "D",
				},
				{
					path:  []string{"S", "Ptr"},
					value: ts.S.Ptr,
				},
				{
					path:  []string{"S", "M", "k"},
					value: 123,
				},
				{
					path:  []string{"Nested", "Val"},
					value: true,
				},
				{
					path:  []string{"Iface", "Val"},
					value: "iface",
				},
				{
					path:  []string{"PtrPtr"},
					value: ts.PtrPtr,
				},
				{
					path:  []string{"M", "key"},
					value: "value",
				},
				{
					path:  []string{"Slice", "0"},
					value: float64(26.05),
				},
				{
					path:  []string{"Slice", "1"},
					value: float64(1.1),
				},
				{
					path:  []string{"Slice", "2"},
					value: float64(23.12),
				},
				{
					path:  []string{"Array", "0"},
					value: float32(1),
				},
				{
					path:  []string{"Array", "1"},
					value: float32(2),
				},
				{
					path:  []string{"Array", "2"},
					value: float32(3),
				},
				{
					path:  []string{"Complex64"},
					value: complex64(complex(1, 2)),
				},
				{
					path:  []string{"Complex128"},
					value: complex(3, 4),
				},
			},
			opts: []Option{
				ExpandUnexported(false),
				AddNilContainers(false),
				AddNilFields(false),
				WithPointerFllowPolicy(PointerPolicyPrimitivePointer),
			},
		},
		{
			obj: 5,
			exp: []pathValue{
				{
					path:  []string{},
					value: 5,
				},
			},
		},
		{
			obj: []int{1, 2, 3},
			exp: []pathValue{
				{
					path:  []string{"0"},
					value: 1,
				},
				{
					path:  []string{"1"},
					value: 2,
				},
				{
					path:  []string{"2"},
					value: 3,
				},
			},
		},
		{
			obj: map[string]string{
				"a": "b",
			},
			exp: []pathValue{
				{
					path:  []string{"a"},
					value: "b",
				},
			},
		},
	}
	for i := range tests {
		idx := i
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			test := tests[idx]
			a := assert.New(t)
			var slice []pathValue
			m := make(map[string]interface{})
			Walk(test.obj, func(path []string, value interface{}) bool {
				p := make([]string, len(path))
				copy(p, path)
				slice = append(slice, pathValue{path: p, value: value})
				m[strings.Join(path, ".")] = value
				return true
			}, test.opts...)
			a.Equal(test.exp, slice)
			a.Equal(m, Flatten(test.obj, test.opts...))
		})
	}
}

type cycleMapTest struct {
	M map[string]cycleMapTest2
}

type cycleMapTest2 struct {
	M map[string]cycleMapTest
}

func TestCyclicMap(t *testing.T) {
	a := assert.New(t)
	m := map[string]cycleMapTest{
		"m": cycleMapTest{
			M: map[string]cycleMapTest2{
				"m2": cycleMapTest2{},
			},
		},
	}
	mm := m["m"]
	mm.M = map[string]cycleMapTest2{
		"m2": cycleMapTest2{
			M: m,
		},
	}
	m["m"] = mm
	a.Equal(map[string]interface{}{}, Flatten(m))
}

func TestSortMapKeys(t *testing.T) {
	m := make(map[string]string)
	var exp, act []pathValue
	for r := 'a'; r <= 'z'; r++ {
		s := string(r)
		m[s] = s
		exp = append(exp, pathValue{
			path:  []string{s},
			value: s,
		})
	}
	Walk(m, func(path []string, value interface{}) bool {
		p := make([]string, len(path))
		copy(p, path)
		act = append(act, pathValue{
			path:  p,
			value: value,
		})
		return true
	}, SortMapKeys(true))
	assert.Equal(t, exp, act)
}

func TestWalkStop(t *testing.T) {
	st := testpkg.NewTestStruct()
	var total int
	opts := []Option{AddNilContainers(true), AddNilFields(true), ExpandUnexported(true)}
	Walk(st, func([]string, interface{}) bool {
		total++
		return true
	}, opts...)
	for i := 0; i < total; i++ {
		var current int
		Walk(st, func([]string, interface{}) bool {
			current++
			return current <= i
		}, opts...)
		assert.Equal(t, i+1, current)
	}
}
