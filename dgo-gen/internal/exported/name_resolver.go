package exported

import (
	"go/types"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
)

type IManager interface {
	GetPackageMapping() *map[string]*Package
	ParsePackages(pkgPath string, isDep bool)
}

type NameResolver struct {
	m IManager
}

func NewNameResolver(m IManager) *NameResolver {
	return &NameResolver{m}
}

func (r *NameResolver) HasTypeName(obj *types.TypeName) bool {
	m := r.m.GetPackageMapping()
	pkgPath := obj.Pkg().Path()
TRY_LOAD_PKG:
	if epkg, pkgExists := (*m)[pkgPath]; pkgExists {
		_, exists := epkg.Types[obj.Name()]
		return exists
	} else {
		r.m.ParsePackages(pkgPath, true)
		goto TRY_LOAD_PKG
	}
}

func (r *NameResolver) ThrowAt(obj types.Object, args ...any) {
	pkgPath := obj.Pkg().Path()
	pkg := (*r.m.GetPackageMapping())[pkgPath].OriPkg
	exception.ThrowAt(pkg, obj, args...)
}
