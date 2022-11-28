package internal

import "errors"

//dgo:export
func (Bool) Func() {}

//dgo:export
func (Bool) FuncReturnsValue() (ret string) {
	return "Hello"
}

//dgo:export
func (Bool) FuncReturnsError() (err error) {
	return errors.New("Hello")
}

//dgo:export
func (Bool) FuncReturnsValueAndError() (ret string, err error) {
	return "", nil
}

//dgo:export
func (Bool) FuncWith1Arg(arg1 float32) {}

//dgo:export
func (Bool) FuncWith2Arg(arg1 float32, arg2 Uint) {}
