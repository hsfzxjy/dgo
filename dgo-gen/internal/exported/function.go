package exported

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal/collector"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"golang.org/x/tools/go/packages"
)

type FunctionParam struct {
	Name string
	Term ir.Term
}

type Function struct {
	*collector.Context
	PPackage        *packages.Package
	Name            string
	ReceiverName    string
	ReceiverPointer bool
	Params          []FunctionParam
	Return          ir.Term
}