package exported

import (
	"go/types"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"golang.org/x/tools/go/packages"
)

type Type struct {
	types.Object
	*NameResolver
	OriPkg   *packages.Package
	resolved bool
	IrTerm   ir.Term
}

func (ex *Type) Resolve() {
	if ex.resolved {
		return
	}
	tr := NewTypeResolver(ex.NameResolver)
	tr.Do(ex)
	ex.IrTerm = tr.Result
}
