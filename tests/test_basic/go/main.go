package main

// #include <stdint.h>
// #include <stdbool.h>
import "C"
import (
	"errors"
	"log"

	dgo "github.com/hsfzxjy/dgo/go"
	_ "github.com/hsfzxjy/dgo/go/pin"
)

func main() {}

func assertEqual(tcase Equalable, other any) {
	if !tcase.Equals(other) {
		log.Panicf("%[1]T %#[1]v != %[2]T %#[2]v", tcase, other)
	} else {
		log.Printf("Test Passed: %[1]T %#[1]v\n", tcase)
	}
}
func resolveCase[T ~uint32](token T) {
	resolveCallback.Flag(dgo.CF).Call(uint32(token))
}

func makeComparator[T Equalable](token uint32, x T) dgo.GoCallback {
	return dgo.Pend(func(t any) {
		if !x.Equals(t) {
			log.Panicf("%[1]T %#[1]v != %[2]T %#[2]v", x, t)
		} else {
			log.Printf("Test Passed: %[1]T %#[1]v\n", x)
		}
		resolveCallback.Flag(dgo.CF).Call(token)
	}, nil)
}

var resolveCallback dgo.DartCallback
var currentPort *dgo.Port
var currentPort2 *dgo.Port

//go:linkname lookupPort github.com/hsfzxjy/dgo/go.lookupPort
func lookupPort(key dgo.PortKey) *dgo.Port

//export InitTestContext
func InitTestContext(cdcb C.uint32_t, portKey C.int64_t, isDefault C.bool) {
	currentPort2 = lookupPort(dgo.PortKey(portKey))
	currentPort = currentPort2
	if isDefault {
		currentPort = nil
	}
	resolveCallback = dgo.WrapDartCallback(cdcb, currentPort)
	groupSingle.reset()
	groupTuple.reset()
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
	dcb := dgo.WrapDartCallback(cdcb, currentPort)

	tcase := groupSingle.next()
	gcb := dgo.Pend(func(other any) {
		assertEqual(tcase, other)
		resolveCase(cdcb)
	}, currentPort)
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
	dcb := dgo.WrapDartCallback(cdcb, currentPort)

	tcase := groupTuple.next()
	gcb := dgo.Pend(func(x1, x2 any) {
		assertEqual(tcase, []any{x1, x2})
		resolveCase(cdcb)
	}, currentPort)
	tcase.SendToDart(func(a ...any) {
		args := []any{gcb}
		args = append(args, a...)
		dcb.Flag(dgo.CF_POP).Call(args...)
	})

	return groupTuple.exhausted()
}

//export TestDartPop
func TestDartPop(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb, currentPort)
	dcb.Flag(dgo.CF).Call()
	dcb.Flag(dgo.CF_POP).Call()
	return 0
}

//export TestDartPackArray
func TestDartPackArray(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb, currentPort)
	dcb.Flag(dgo.CF_POP.PackArray()).Call(1, "hello", 3.14)
	return 0
}

//export TestDartWithContext
func TestDartWithContext(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb, currentPort)
	dcb.Flag(dgo.CF_POP.WithContext()).Call(1, "hello", 3.14)
	return 0
}

//export TestDartFast
func TestDartFast(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb, currentPort)
	dcb.Flag(dgo.CF_FAST_NIL).Call()
	dcb.Flag(dgo.CF_FAST_NO).Call()
	dcb.Flag(dgo.CF_FAST_YES.Pop()).Call()

	return 0
}

//export TestDartFastVoid
func TestDartFastVoid(cdcb C.uint32_t) int {
	dcb := dgo.WrapDartCallback(cdcb, currentPort)
	dcb.Flag(dgo.CF_FAST_VOID.Pop()).Call()

	return 0
}

//export TestDartFallible
func TestDartFallible(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb, currentPort).Flag(dgo.CF_POP.Fallible()).Call()
	return 0
}

//export TestDartFutureResolve
func TestDartFutureResolve(cdcb C.uint32_t) int {
	dgo.DartFutureCallback(dgo.WrapDartCallback(cdcb, currentPort)).Resolve(42)
	return 0
}

//export TestDartFutureReject
func TestDartFutureReject(cdcb C.uint32_t) int {
	dgo.DartFutureCallback(dgo.WrapDartCallback(cdcb, currentPort)).Reject(errors.New("this is an error"))
	return 0
}

//export TestDartStream
func TestDartStream(cdcb C.uint32_t) int {
	dcb := dgo.DartStreamCallback(dgo.WrapDartCallback(cdcb, currentPort))
	dcb.Add(1)
	dcb.Add(3.14)
	dcb.AddError(errors.New("error 1"))
	dcb.Add("4")
	dcb.Done()
	return 0
}

//export TestGoPop
func TestGoPop(cdcb C.uint32_t) int {
	counter := 0
	var gcb dgo.GoCallback
	gcb = dgo.Pend(func() {
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
	}, currentPort)
	dgo.WrapDartCallback(cdcb, currentPort).Flag(dgo.CF.Pop()).Call(gcb)
	return 0
}

//export TestGoWithContext
func TestGoWithContext(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb, currentPort).
		Flag(dgo.CF.Pop()).
		Call(dgo.Pend(func(context *dgo.InvokeContext, a1 int64, a2 string, a3 float64, a4 any) {
			cf := context.Flag()
			if !cf.HasPop() || !cf.HasWithContext() ||
				a1 != 1 || a2 != "hello" || a3 != 3.14 || a4 != nil {
				panic("failed")
			}
			resolveCase(cdcb)
		}, currentPort))
	return 0
}

//export TestGoPackArray
func TestGoPackArray(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb, currentPort).
		Flag(dgo.CF.Pop()).
		Call(dgo.Pend(func(arr []any) {
			cf := arr[0].(*dgo.InvokeContext).Flag()
			a1 := arr[1].(int64)
			a2 := arr[2].(string)
			a3 := arr[3].(float64)
			if !cf.HasPop() || !cf.HasWithContext() || !cf.HasPackArray() ||
				a1 != 1 || a2 != "hello" || a3 != 3.14 {
				panic("failed")
			}
			resolveCase(cdcb)
		}, currentPort))
	return 0
}

//export TestGoFast
func TestGoFast(cdcb C.uint32_t) int {
	counter := 0
	dgo.WrapDartCallback(cdcb, currentPort).
		Flag(dgo.CF.Pop()).
		Call(dgo.Pend(func(ans any) {
			if ans != []any{nil, false, true}[counter] {
				log.Panicf("failed, counter = %v, ans = %v", counter, ans)
			}
			counter++
			if counter == 3 {
				resolveCase(cdcb)
			}
		}, currentPort))
	return 0
}

//export TestGoFastVoid
func TestGoFastVoid(cdcb C.uint32_t) int {
	dgo.WrapDartCallback(cdcb, currentPort).
		Flag(dgo.CF.Pop()).
		Call(dgo.Pend(func() {
			resolveCase(cdcb)
		}, currentPort))
	return 0
}

//export TestGoFallible
func TestGoFallible(cdcb C.uint32_t) int {
	gcb := dgo.Pend(func() {}, currentPort)
	gcb.Remove()
	dgo.WrapDartCallback(cdcb, currentPort).
		Flag(dgo.CF.Pop()).
		Call(gcb)
	return 0
}

//go:linkname isPortGone github.com/hsfzxjy/dgo/go.isPortGone
func isPortGone(key dgo.PortKey, isDefault bool) bool

//go:linkname testPostToDartPort github.com/hsfzxjy/dgo/go.testPostToDartPort
func testPostToDartPort(port *dgo.Port) bool

//go:linkname testPostToGoPort github.com/hsfzxjy/dgo/go.testPostToGoPort
func testPostToGoPort(port *dgo.Port) bool

//export TestPortClosed
func TestPortClosed(key C.int64_t, isDefault bool) {
	if !isPortGone(dgo.PortKey(key), isDefault) {
		log.Panicf("port still exists, key=%d, isDefault=%v", key, isDefault)
	}
	if testPostToDartPort(currentPort2) {
		panic("dart port still alive")
	}
	if testPostToGoPort(currentPort2) {
		panic("go port still alive")
	}
}
