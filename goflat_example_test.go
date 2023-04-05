package goflat

import (
	"fmt"
	"strings"

	"github.com/avdva/goflat/testpkg"
)

func ExampleWalk() {
	ts := testpkg.NewTestStruct()
	Walk(ts, func(path []string, value interface{}) bool {
		fmt.Printf("%q ---> %v\n", strings.Join(path, "."), value)
		return true
	}, SortMapKeys(true), ExpandUnexported(true), WithPointerFllowPolicy(PointerPolicyJustValue))
	// Output:
	//"A" ---> 5
	//"B" ---> 6
	//"S.D" ---> D
	//"S.Ptr" ---> 123
	//"S.M.k" ---> 123
	//"S.notExportedMap.k" ---> v
	//"Nested.Val" ---> true
	//"embedded.S" ---> 123
	//"Iface.Val" ---> iface
	//"PtrPtr" ---> 123
	//"M.key" ---> value
	//"Slice.0" ---> 26.05
	//"Slice.1" ---> 1.1
	//"Slice.2" ---> 23.12
	//"Array.0" ---> 1
	//"Array.1" ---> 2
	//"Array.2" ---> 3
	//"Complex64" ---> (1+2i)
	//"Complex128" ---> (3+4i)
	//"notExportedInt" ---> 123
	//"notExportedIface.Val" ---> string
	//"notExportedStruct.A" ---> 0
	//"notExportedPointer" ---> string
}
