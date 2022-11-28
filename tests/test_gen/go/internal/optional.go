package internal

import dgo "github.com/hsfzxjy/dgo/go"

//dgo:export
type OptionalWithZeroSizedContent struct {
	field dgo.Optional[[0]int]
}
