package main

import "errors"

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

func main() {}
