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

opts := []Option{
	ExpandUnexported(true),                    // go inside unexported fields
	AddNilContainers(true),                    // include nil maps/slices
	AddNilFields(true),                        // include nil pointers to primitive types
	SortMapKeys(true),                         // sort map keys before visiting a map
	WithPointerFllowPolicy(PointerPolicyBoth), // how to handle pointers. see PointerPolicy* consts.
}

// Walk will call the the callback with corresponding path and value
// for all objects inside the root value object.
Walk(object, func(path []string, value interface{}) {}, opts...)

// Flatten will build a map from element's path to a value.
Flatten(object, opts...)
```

## Example
The following struct
```go

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

```

flattens to:
```
"A" ---> 5
"B" ---> 6
"S.D" ---> D
"S.Ptr" ---> 123
"S.M.k" ---> 123
"S.notExportedMap.k" ---> v
"Nested.Val" ---> true
"embedded.S" ---> 123
"Iface.Val" ---> iface
"PtrPtr" ---> 123
"M.key" ---> value
"Slice.0" ---> 26.05
"Slice.1" ---> 1.1
"Slice.2" ---> 23.12
"Array.0" ---> 1
"Array.1" ---> 2
"Array.2" ---> 3
"Complex64" ---> (1+2i)
"Complex128" ---> (3+4i)
"notExportedInt" ---> 123
"notExportedIface.Val" ---> string
"notExportedStruct.A" ---> 0
"notExportedPointer" ---> string
```

Please see the [examples](/goflat_example_test.go) for more details.

## Contact

[Aleksandr Demakin](mailto:alexander.demakin@gmail.com)

## License

Source code is available under the [Apache License Version 2.0](/LICENSE).