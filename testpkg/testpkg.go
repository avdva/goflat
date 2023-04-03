package testpkg

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

func NewTestStruct() *testStruct {
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
		notExportedStruct: struct {
			A int
		}{
			A: 0,
		},
		notExportedPointer: &s,
	}
}

type SmallStruct struct {
	Data *float64
}

type PointerTestStruct struct {
	IntPtr              *int
	IntPtrPtr           **int
	unexportedString    *string
	unexportedNil       *string
	ExportedStruct      *SmallStruct
	unexportedStruct    *SmallStruct
	unexportedNilStruct *SmallStruct
	unexportedSlice     []int
}

func NewPointerTestStruct() *PointerTestStruct {
	intPtr := new(int)
	*intPtr = 5
	s := "hello"
	flt := 123.456
	return &PointerTestStruct{
		IntPtr:           intPtr,
		IntPtrPtr:        &intPtr,
		unexportedString: &s,
		unexportedNil:    nil,
		ExportedStruct: &SmallStruct{
			Data: &flt,
		},
		unexportedStruct: &SmallStruct{
			Data: &flt,
		},
		unexportedSlice: []int{1, 2, 3},
	}
}
