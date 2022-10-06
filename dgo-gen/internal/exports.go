package internal

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal/dumper"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/dumper/go_dumper"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/dumper/ir_dumper"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
)

type Exports struct {
	Packages     map[string]*exported.Package
	NameResolver *exported.NameResolver
}

func NewExports(pkgPath string) *Exports {
	ex := &Exports{Packages: make(map[string]*exported.Package)}
	ex.NameResolver = exported.NewNameResolver(ex)
	ex.ParsePackages(pkgPath, false)
	return ex
}

func (ex *Exports) GetPackageMapping() *map[string]*exported.Package { return &ex.Packages }
func (ex *Exports) ParsePackages(pkgPath string, isDep bool) {
	pkgs := LoadPackages(pkgPath)
	for _, pkg := range pkgs {
		ex.Packages[pkg.PkgPath] = exported.NewPackage(pkg, ex.NameResolver, isDep)
	}
	if !isDep {
		for _, pkg := range pkgs {
			ex.Packages[pkg.PkgPath].Resolve()
		}
	}
}

func (ex *Exports) Dump() {
	dumpers := []dumper.Dumper{
		&go_dumper.GoDumper{},
		&ir_dumper.IrDumper{},
	}
	for _, dumper := range dumpers {
		for _, epkg := range ex.Packages {
			if epkg.IsDep {
				continue
			}
			for _, etype := range epkg.Types {
				dumper.AddType(etype)
			}
		}
		dumper.Save()
	}
}
