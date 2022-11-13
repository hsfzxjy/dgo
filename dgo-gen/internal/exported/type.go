package exported

import (
	"go/types"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/collector"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
	"golang.org/x/tools/go/packages"
)

type TypeMethod struct {
	Name string
	*Function
}

type Type struct {
	types.Object
	*collector.Context
	PPackage *packages.Package
	Term     ir.Term
	TypeId   int64
	Methods  []TypeMethod
}

func (t *Type) Uri() uri.Uri {
	return uri.UriFor(t.PPackage, t.Name())
}

func (t *Type) Resolve() {
	t.Term = t.SolveType(t)
}
