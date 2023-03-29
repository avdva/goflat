package goflat

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	A int
	B uint64
	S struct {
		D              string
		Ptr            *int
		M              map[string]int
		notExportedMap map[string]string
	}
	Nested nested
	embedded
	Iface             privateInterface
	PtrPtr            **int
	M                 *map[string]string
	Slice             []float64
	Array             [3]float32
	NilSlice          []int
	Complex64         complex64
	Complex128        complex128
	notExportedInt    int
	notExportedIface  privateInterface
	notExportedStruct struct {
		A int
	}
	notExportedPointer *string
}

type embedded struct {
	S int8
}

type nested struct {
	Val bool
}

type privateInterface interface {
	f()
}

type interfaceImpl struct {
	Val string
}

func (impl *interfaceImpl) f() {}

func newTestStruct() *testStruct {
	i := 123
	ptr := &i
	ss := map[string]string{
		"key": "value",
	}
	s := "string"
	return &testStruct{
		A: 5,
		B: 6,
		S: struct {
			D              string
			Ptr            *int
			M              map[string]int
			notExportedMap map[string]string
		}{
			D:   "D",
			Ptr: ptr,
			M: map[string]int{
				"k": 123,
			},
			notExportedMap: map[string]string{
				"k": "v",
			},
		},
		Nested: nested{
			Val: true,
		},
		embedded: embedded{
			S: 123,
		},
		Iface: &interfaceImpl{
			Val: "iface",
		},
		PtrPtr: &ptr,
		M:      &ss,
		Slice: []float64{
			26.05, 1.1, 23.12,
		},
		Array: [3]float32{
			1, 2, 3,
		},
		Complex64:      complex(1, 2),
		Complex128:     complex(3, 4),
		notExportedInt: 123,
		notExportedIface: &interfaceImpl{
			Val: s,
		},
		notExportedPointer: &s,
	}
}

type pathValue struct {
	path  []string
	value interface{}
}

func TestFlatten(t *testing.T) {
	ts := newTestStruct()

	tests := []struct {
		obj  interface{}
		exp  []pathValue
		opts []Option
	}{
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
			},
			opts: []Option{
				ExpandUnexported(true),
				AddNilContainers(true),
				AddNilFields(true),
				SortMapKeys(true),
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
			},
		},
		{
			obj: 5,
			exp: []pathValue{
				{
					path:  []string{"."},
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
