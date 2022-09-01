package main

// #include <stdint.h>
import "C"
import (
	"dgo"
	"errors"
	"fmt"
)

func main() {}

func assertEqual(tcase Equalable, other any) {
	if !tcase.Equals(other) {
		panic(fmt.Sprintf("%[1]T %#[1]v != %[2]T %#[2]v", tcase, other))
	} else {
		fmt.Printf("Test Passed: %[1]T %#[1]v\n", tcase)
	}
}
func resolveCase[T ~uint32](token T) {
	resolveCallback.Flag(dgo.CF).Call(uint32(token))
}

func makeComparator[T Equalable](token uint32, x T) dgo.GoCallback {
	return dgo.PendGo(func(t any) {
		if !x.Equals(t) {
			panic(fmt.Sprintf("%[1]T %#[1]v != %[2]T %#[2]v", x, t))
		} else {
			fmt.Printf("Test Passed: %[1]T %#[1]v\n", x)
		}
		resolveCallback.Flag(dgo.CF).Call(token)
	})
}

var resolveCallback dgo.DartCallback

//export SetResolveCallback
func SetResolveCallback(cdcb C.uint32_t) {
	resolveCallback = dgo.WrapDartCallback(cdcb)
}

var sampleByteSlice = []byte("This is a slice")

var groupSingle = group([]Equalable{
	vs[bool](true),
	vc[uint, int64](^uint(0)),
	vc[uint8, int64](^uint8(0)),
	vc[uint16, int64](^uint16(0)),
	vc[uint32, int64](^uint32(0)),
	vc[uint64, int64](^uint64(0)),
	vc[int, int64](^int(0)),
	vc[int8, int64](^int8(0)),
	vc[int16, int64](^int16(0)),
	vc[int32, int64](^int32(0)),
	vc[int64, int64](^int64(0)),
	vc[float32, float64](3.14),
	vs[float64](3.14),
	vs[string]("hello world"),
	slice[byte](sampleByteSlice),
	eslice[byte](sampleByteSlice),
})

//export TestSingle
func TestSingle(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb)

	tcase := groupSingle.next()
	gcb := dgo.PendGo(func(other any) {
		assertEqual(tcase, other)
		resolveCase(cdcb)
	})
	tcase.SendToDart(func(a ...any) {
		args := []any{gcb}
		args = append(args, a...)
		dcb.Flag(dgo.CF_POP).Call(args...)
	})

	return groupSingle.exhausted()
}

var groupTuple = group([]Equalable{
	tc[bool, uint8, bool, int64](true, ^uint8(0)),
	ts[string, int64]("Hi", ^int64(0)),
})

//export TestTuple
func TestTuple(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb)

	tcase := groupTuple.next()
	gcb := dgo.PendGo(func(x1, x2 any) {
		assertEqual(tcase, []any{x1, x2})
		resolveCase(cdcb)
	})
	tcase.SendToDart(func(a ...any) {
		args := []any{gcb}
		args = append(args, a...)
		dcb.Flag(dgo.CF_POP).Call(args...)
	})

	return groupTuple.exhausted()
}

//export TestDartPop
func TestDartPop(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb)
	dcb.Flag(dgo.CF).Call()
	dcb.Flag(dgo.CF_POP).Call()
	return 0
}

//export TestDartPackArray
func TestDartPackArray(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb)
	dcb.Flag(dgo.CF_POP.PackArray()).Call(1, "hello", 3.14)
	return 0
}

//export TestDartWithCode
func TestDartWithCode(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb)
	dcb.Flag(dgo.CF_POP.WithCode()).Call(1, "hello", 3.14)
	return 0
}

//export TestDartFast
func TestDartFast(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb)
	dcb.Flag(dgo.CF_FAST_NIL).Call()
	dcb.Flag(dgo.CF_FAST_NO).Call()
	dcb.Flag(dgo.CF_FAST_YES.Pop()).Call()

	return 0
}

//export TestDartFastVoid
func TestDartFastVoid(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb)
	dcb.Flag(dgo.CF_FAST_VOID.Pop()).Call()

	return 0
}

//export TestDartFutureResolve
func TestDartFutureResolve(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb).AsFut().Resolve(42)
	return 0
}

//export TestDartFutureReject
func TestDartFutureReject(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb).AsFut().Reject(errors.New("this is an error"))
	return 0
}

//export TestGoPop
func TestGoPop(cdcb C.uint32_t) int {
	counter := 0
	var gcb dgo.GoCallback
	gcb = dgo.PendGo(func() {
		counter++
		if counter == 1 {
			if !gcb.Exists() {
				panic("should exist")
			}
		}
		if counter == 2 {
			if gcb.Exists() {
				panic("should not exist")
			}
			resolveCase(cdcb)
		}
	})
	dgo.WrapDartCallback(cdcb).Flag(dgo.CF.Pop()).Call(gcb)
	return 0
}

//export TestGoWithCode
func TestGoWithCode(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb).
		Flag(dgo.CF.Pop()).
		Call(dgo.PendGo(func(cf dgo.CallbackFlag, a1 int64, a2 string, a3 float64, a4 any) {
			if !cf.HasPop() || !cf.HasWithCode() ||
				a1 != 1 || a2 != "hello" || a3 != 3.14 || a4 != nil {
				panic("failed")
			}
			resolveCase(cdcb)
		}))
	return 0
}

//export TestGoPackArray
func TestGoPackArray(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb).
		Flag(dgo.CF.Pop()).
		Call(dgo.PendGo(func(arr []any) {
			cf := arr[0].(dgo.CallbackFlag)
			a1 := arr[1].(int64)
			a2 := arr[2].(string)
			a3 := arr[3].(float64)
			if !cf.HasPop() || !cf.HasWithCode() || !cf.HasPackArray() ||
				a1 != 1 || a2 != "hello" || a3 != 3.14 {
				panic("failed")
			}
			resolveCase(cdcb)
		}))
	return 0
}

//export TestGoFast
func TestGoFast(cdcb C.uint32_t) int {
	counter := 0
	dgo.WrapDartCallback(cdcb).
		Flag(dgo.CF.Pop()).
		Call(dgo.PendGo(func(ans any) {
			if ans != []any{nil, false, true}[counter] {
				panic(fmt.Sprintf("failed, counter = %v, ans = %v", counter, ans))
			}
			counter++
			if counter == 3 {
				resolveCase(cdcb)
			}
		}))
	return 0
}

//export TestGoFastVoid
func TestGoFastVoid(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb).
		Flag(dgo.CF.Pop()).
		Call(dgo.PendGo(func() {
			resolveCase(cdcb)
		}))
	return 0
}
