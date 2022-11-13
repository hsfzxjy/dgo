package exported

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal/collector"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/interfaces"
	"golang.org/x/tools/go/packages"
)

type Package struct {
	_PPackage *packages.Package
	Types     []*Type
	TypeNames map[string]int
	Funcs     []*Function
	IsDep     bool
}

func NewPackage(ppkg *packages.Package, ctx *collector.Context, isDep bool) *Package {
	epkg := &Package{
		_PPackage: ppkg,
		TypeNames: make(map[string]int),
		IsDep:     isDep,
	}

	ctx.Collect(
		ppkg,
		func(typeName string) {
			object := ppkg.Types.Scope().Lookup(typeName)
			epkg.TypeNames[typeName] = len(epkg.Types)
			epkg.Types = append(epkg.Types, &Type{
				Context:  ctx,
				Object:   object,
				PPackage: ppkg,
				TypeId:   ctx.NextTypeId(),
			})
		},
		func(funcName, recvName string, recvPointer bool) {
			epkg.Funcs = append(epkg.Funcs, &Function{
				Context:         ctx,
				PPackage:        ppkg,
				Name:            funcName,
				ReceiverName:    recvName,
				ReceiverPointer: recvPointer,
				FuncId:          ctx.NextMethodId(),
			})
		},
	)

	return epkg
}

func (epkg *Package) Resolve() {
	for _, etyp := range epkg.Types {
		etyp.Resolve()
	}

	for _, efunc := range epkg.Funcs {
		efunc.Resolve()
	}
	epkg.Funcs = nil
}

func (epkg *Package) HasType(name string) bool {
	return epkg.Type(name) != nil
}

func (epkg *Package) Type(name string) interfaces.Type {
	if id, ok := epkg.TypeNames[name]; ok {
		return epkg.Types[id]
	}
	return nil
}

func (epkg *Package) PPackage() *packages.Package { return epkg._PPackage }
