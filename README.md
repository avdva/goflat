# goflat
The package allows flattening of complex Go objects.

## Badges

![Build Status](https://github.com/avdva/goflat/workflows/golangci-lint/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/avdva/goflat)](https://goreportcard.com/report/github.com/avdva/goflat)

## Installation

To start using this package, run:

```sh
$ go get github.com/avdva/goflat
```

## API

```go

opts := ExpandUnexported(true), AddNilContainers(true), AddNilFields(true), SortMapKeys(true)

// Walk will call the callback with corresponding path and value for all objects of primitive types of
// structs, pointers, arrays, slices and maps(those having string keys).
Walk(object, func(path []string, value interface{}) {}, opts...)

// Flatten will build a map from element's path to a value.
Flatten(object, opts...)
```

## Example
The following structure
```go

type testStruct struct {
	A int
	B uint64
	S struct {
		D   string
		Ptr *int
		M   map[string]int
	}
	Nested nested
	embedded
	Iface  privateInterface
	PtrPtr **int
	M      *map[string]string
	Slice  []float64
	Array  [3]float32
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
			D   string
			Ptr *int
			M   map[string]int
		}{
			D:   "D",
			Ptr: ptr,
			M: map[string]int{
				"k": 123,
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
```

flattens to:
```
"A" --> 5
"B" --> 6
"S.D" --> D
"S.Ptr" --> 0xc0000ca000
"S.M.k": --> 123
"Nested.Val" --> true
"embedded.S" --> 123
"Iface.Val" --> iface
"PtrPtr" --> 0xc0000c4018
"M.key" --> value
"Slice.0" --> 26.05
"Slice.1" --> 1.1
"Slice.2" --> 23.12
"Array.0" --> 1
"Array.1" --> 2
"Array.2" --> 3
```
