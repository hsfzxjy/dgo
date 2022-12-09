package main

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"

	"github.com/hsfzxjy/dgo/go/pin"
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
}

//dgo:export
type PinTester struct{}

//dgo:export
func (PinTester) MakeAndReturnsPeripheral() pin.Token[Peripheral] {
	// we allocate a large array to ensure the garbadge collector will actively
	// recycle it when runtime.GC() called
	peripherals := new([100_0000]Peripheral)
	p := &peripherals[0]
	p.id = 42
	p.name = "MyDevice"
	p.Pin()
	t := p.NewToken()
	if !t.Dispose() || t.Dispose() {
		panic("expect 1-st Dispose() returns true, and the 2-nd returns false")
	}
	t = p.NewToken()
	return t
}

//dgo:export
func (PinTester) AcceptPeripheralAndCompute(pt pin.Token[Peripheral]) string {
	p := pt.Data()
	p.Unpin()
	return fmt.Sprintf("Peripheral<id=%d, name=%s>", p.id, p.name)
}

//dgo:export
func (PinTester) GC() {
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
