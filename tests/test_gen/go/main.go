package main

import (
	"errors"

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
func (Tester) ReturnsExternalType() subpack.StructInSubpack {
	return subpack.StructInSubpack{FieldString: "string"}
}

func main() {}
