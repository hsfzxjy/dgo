package main

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"time"

	dgo "github.com/hsfzxjy/dgo/go"
	"github.com/hsfzxjy/dgo/go/pin"
	"github.com/hsfzxjy/dgo/tests/test_gen/go/internal"
	"github.com/hsfzxjy/dgo/tests/test_gen/go/internal/subpack"
)

//dgo:export
type Tester struct{}

//dgo:export
func (Tester) ReturnsVoid() {}

//dgo:export
func (Tester) ReturnsString() string { return "Hello world!" }

//dgo:export
func (Tester) ReturnsError() error { return errors.New("error") }

//dgo:export
func (Tester) ReturnsStringOrError(success bool) (string, error) {
	if success {
		return "success", nil
	} else {
		return "", errors.New("error")
	}
}

//dgo:export
func (Tester) ReturnsSlice(n int) (ret []string) {
	for i := 0; i < n; i++ {
		ret = append(ret, strconv.Itoa(i))
	}
	return
}

//dgo:export
func (Tester) ReturnsIdentitySlice(x []int) []int {
	return x
}

//dgo:export
func (Tester) ReturnsIdentityMap(x map[int]string) map[int]string {
	return x
}

//dgo:export
func (Tester) ReturnsMap(n int) (ret map[int]string) {
	ret = make(map[int]string)
	for i := 0; i < n; i++ {
		ret[i] = strconv.Itoa(i)
	}
	return
}

//dgo:export
func (Tester) ReturnsExternalType() subpack.StructInSubpack {
	return subpack.StructInSubpack{FieldString: "string"}
}

//dgo:export
func (Tester) ReturnsStructWithSimpleTypes(s internal.StructWithSimpleTypes) internal.StructWithSimpleTypes {
	return s
}

//dgo:export
type TesterWithField struct {
	field int
}

//dgo:export
func (t TesterWithField) ReturnsSelf() TesterWithField {
	return t
}

//dgo:export ,pinnable
type Peripheral struct {
	pin.Meta
	id   int64
	name string

	state          chan int
	stateBlock     chan int `dgo:",block"`
	stateBroadcast chan int `dgo:",broadcast"`
	stateMemo      chan int `dgo:",memo"`
}

//dgo:export
func (p *Peripheral) ToString() string {
	return fmt.Sprintf("Peripheral<id=%d, name=%s>", p.id, p.name)
}

//dgo:export
type PinTester struct{}

//dgo:export
func (PinTester) MakeAndReturnsPeripheral(rcb uint32) pin.Token[Peripheral] {
	cb := dgo.DartFutureCallback(dgo.WrapDartCallback(rcb, nil))
	// we allocate a large array to ensure the garbadge collector will actively
	// recycle it when runtime.GC() called
	peripherals := new([100_0000]Peripheral)
	runtime.SetFinalizer(peripherals, func(_ any) {
		cb.Resolve(true)
	})
	p := &peripherals[0]
	p.id = 42
	p.name = "MyDevice"
	p.state = make(chan int)
	p.stateBlock = make(chan int)
	p.stateBroadcast = make(chan int)
	p.stateMemo = make(chan int)
	p.Pin()
	t := p.NewToken()
	if !t.Dispose() || t.Dispose() {
		panic("expect 1-st Dispose() returns true, and the 2-nd returns false")
	}
	t = p.NewToken()
	return t
}

func send(ch chan<- int, value int) {
	select {
	case ch <- value:
	case <-time.After(200 * time.Millisecond):
		println("send timeout")
	}
}

//dgo:export
func (PinTester) StartStateAndUnpin(t pin.Token[Peripheral], async_ bool, toClose bool) {
	data := t.Data()
	f := func() {
		send(data.state, 1)
		send(data.state, 2)
		send(data.state, 3)
		if toClose {
			close(data.state)
		}
	}
	if async_ {
		go f()
	} else {
		f()
	}
	data.Unpin()
}

//dgo:export
func (PinTester) StartStateBroadcastAndUnpin(t pin.Token[Peripheral]) {
	data := t.Data()
	go func() {
		send(data.stateBroadcast, 1)
		send(data.stateBroadcast, 2)
		send(data.stateBroadcast, 3)
		close(data.stateBroadcast)
	}()
	data.Unpin()
}

//dgo:export
func (PinTester) StartStateBlockAndUnpin(t pin.Token[Peripheral]) {
	data := t.Data()
	go func() {
		send(data.stateBlock, 1)
		send(data.stateBlock, 2)
		send(data.stateBlock, 3)
		close(data.stateBlock)
	}()
	data.Unpin()
}

//dgo:export
func (PinTester) StartStateMemoAndUnpin(t pin.Token[Peripheral]) {
	data := t.Data()
	send(data.stateMemo, 1)
	send(data.stateMemo, 2)
	send(data.stateMemo, 3)
	data.Unpin()
}

//dgo:export
func (PinTester) AcceptPeripheralAndCompute(pt pin.Token[Peripheral]) string {
	p := pt.Data()
	p.Unpin()
	return fmt.Sprintf("Peripheral<id=%d, name=%s>", p.id, p.name)
}

//dgo:export
func (PinTester) GC() {
	// invoke runtime.GC() twice to ensure the finalizer is called
	runtime.GC()
	runtime.GC()
}

//dgo:export
func (PinTester) AssertTokenInvalid(t pin.Token[Peripheral]) {
	if !t.IsEmpty() {
		panic("expect the token to be invalid")
	}
	assertPanic(func() { t.Data() }, regexp.MustCompile(`^dgo:go:.*empty Token$`))
}

func assertPanic(f func(), errPattern *regexp.Regexp) {
	defer func() {
		p, ok := recover().(string)
		if !ok {
			panic("panic payload is not string")
		}
		if !errPattern.MatchString(p) {
			panic(fmt.Sprintf(
				"panic payload does not match the desired pattern"+":\n payload=%q\n pattern=%q",
				p, errPattern))
		}
	}()
	f()
}

func main() {}
