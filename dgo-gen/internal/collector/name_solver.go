package collector

import (
	"go/types"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/interfaces"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type NameSolver struct {
	m interfaces.Storage
}

func NewNameSolver(m interfaces.Storage) *NameSolver {
	return &NameSolver{m}
}

func (r *NameSolver) HasTypeName(obj *types.TypeName) bool {
	pkg := r.PackageOf(obj.Pkg().Path())
	return pkg.HasType(obj.Name())
}

func (r *NameSolver) ThrowAt(obj types.Object, args ...any) {
	pkgPath := obj.Pkg().Path()
	r.m.ParsePackages(pkgPath, true)
	ppkg := r.m.Package(pkgPath).PPackage()
	exception.ThrowAt(ppkg, obj, args...)
}

func (r *NameSolver) PackageOf(pkgPath string) interfaces.Package {
TRY_LOAD_PKG:
	if pkg := r.m.Package(pkgPath); pkg != nil {
		return pkg
	} else {
		r.m.ParsePackages(pkgPath, true)
		goto TRY_LOAD_PKG
	}

}

func (r *NameSolver) UriFor(obj types.Object) uri.Uri {
	ppkg := r.PackageOf(obj.Pkg().Path()).PPackage()
	return uri.UriFor(ppkg, obj.Name())
}
