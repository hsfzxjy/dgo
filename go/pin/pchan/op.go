package pchan

import (
	dgo "github.com/hsfzxjy/dgo/go"
	"github.com/hsfzxjy/dgo/go/pin/pcop"
)

type OpKind = pcop.OpKind
type Op = pcop.Op

func portOf(op Op) *dgo.Port                { return (*dgo.Port)(op.Port) }
func AsDartCallback(op Op) dgo.DartCallback { return dgo.WrapDartCallback(op.Dcb, portOf(op)) }
