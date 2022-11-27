package goflat

import (
	"strconv"
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
	}
}

type cycleMapTest struct {
	m map[string]cycleMapTest2
}

type cycleMapTest2 struct {
	m map[string]cycleMapTest
}

func TestFlatten(t *testing.T) {
	ts := newTestStruct()
	a := assert.New(t)
	m := map[string]cycleMapTest{
		"m": cycleMapTest{
			m: map[string]cycleMapTest2{
				"m2": cycleMapTest2{},
			},
		},
	}
	mm := m["m"]
	mm.m = map[string]cycleMapTest2{
		"m2": cycleMapTest2{
			m: m,
		},
	}
	m["m"] = mm
	tests := []struct {
		obj  interface{}
		exp  map[string]interface{}
		opts []Option
	}{
		{
			obj: ts,
			exp: map[string]interface{}{
				"A":                  5,
				"B":                  uint64(6),
				"S.D":                "D",
				"S.Ptr":              ts.S.Ptr,
				"S.M.k":              123,
				"Nested.Val":         true,
				"embedded.S":         int8(123),
				"Iface.Val":          "iface",
				"PtrPtr":             ts.PtrPtr,
				"M.key":              "value",
				"Slice.0":            float64(26.05),
				"Slice.1":            float64(1.1),
				"Slice.2":            float64(23.12),
				"NilSlice.":          nil,
				"Array.0":            float32(1),
				"Array.1":            float32(2),
				"Array.2":            float32(3),
				"notExportedPointer": nil,
			},
			opts: []Option{
				ExpandUnexported(true),
			},
		},
		{
			obj: ts,
			exp: map[string]interface{}{
				"A":          5,
				"B":          uint64(6),
				"S.D":        "D",
				"S.Ptr":      ts.S.Ptr,
				"S.M.k":      123,
				"Nested.Val": true,
				"Iface.Val":  "iface",
				"PtrPtr":     ts.PtrPtr,
				"M.key":      "value",
				"Slice.0":    float64(26.05),
				"Slice.1":    float64(1.1),
				"Slice.2":    float64(23.12),
				"NilSlice.":  nil,
				"Array.0":    float32(1),
				"Array.1":    float32(2),
				"Array.2":    float32(3),
			},
			opts: []Option{
				ExpandUnexported(false),
			},
		},
		{
			obj: 5,
			exp: map[string]interface{}{
				"int": 5,
			},
		},
		{
			obj: m,
			exp: map[string]interface{}{},
		},
		{
			obj: []int{1, 2, 3},
			exp: map[string]interface{}{
				"0": 1,
				"1": 2,
				"2": 3,
			},
		},
		{
			obj: map[string]string{
				"a": "b",
			},
			exp: map[string]interface{}{
				"a": "b",
			},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a.Equal(test.exp, Flatten(test.obj, test.opts...))
		})
	}
}
