package internal

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal/collector"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/generator"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/generator/gogen"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/generator/irgen"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/interfaces"
	"golang.org/x/tools/go/packages"
)

type Exports struct {
	Packages map[string]*exported.Package
	Context  *collector.Context
}

func NewExports(pkgPath string) *Exports {
	ex := &Exports{Packages: make(map[string]*exported.Package)}
	ex.Context = collector.NewContext(ex)
	ex.ParsePackages(pkgPath, false)
	return ex
}

func (ex *Exports) Package(name string) interfaces.Package { return ex.Packages[name] }

func (ex *Exports) ParsePackages(pkgPath string, isDep bool) {
	pkgs := LoadPackages(pkgPath)
	missingPkgs := []*packages.Package{}
	for _, pkg := range pkgs {
		if _, ok := ex.Packages[pkg.PkgPath]; !ok {
			missingPkgs = append(missingPkgs, pkg)
		}
	}
	for _, pkg := range missingPkgs {
		ex.Packages[pkg.PkgPath] = exported.NewPackage(pkg, ex.Context, isDep)
	}
	if !isDep {
		for _, pkg := range missingPkgs {
			ex.Packages[pkg.PkgPath].Resolve()
		}
	}
}

func (ex *Exports) Dump() {
	generators := []generator.Interface{
		&gogen.Generator{},
		&irgen.Generator{},
	}
	for _, generator := range generators {
		for _, epkg := range ex.Packages {
			if epkg.IsDep {
				continue
			}
			for _, etype := range epkg.Types {
				generator.AddType(etype)
			}
		}
		generator.Save()
	}
}
