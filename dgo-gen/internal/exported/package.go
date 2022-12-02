package exported

import (
	"go/constant"
	"go/token"
	"go/types"
	"strconv"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/collector"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/interfaces"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"golang.org/x/tools/go/packages"
)

type Package struct {
	_PPackage *packages.Package
	Types     []*Type
	TypeNames map[string]int
	Funcs     []*Function
	IsDep     bool

	constantDirectives map[string][]string
}

func NewPackage(ppkg *packages.Package, ctx *collector.Context, isDep bool) *Package {
	epkg := &Package{
		_PPackage:          ppkg,
		TypeNames:          make(map[string]int),
		IsDep:              isDep,
		constantDirectives: make(map[string][]string),
	}

	ctx.Collect(
		ppkg,
		func(typeName string, directives []string) {
			object := ppkg.Types.Scope().Lookup(typeName)
			epkg.TypeNames[typeName] = len(epkg.Types)
			t := &Type{
				Context:     ctx,
				Object:      object,
				PPackage:    ppkg,
				TypeId:      ctx.NextTypeId(),
				Methods:     []TypeMethod{},
				EnumMembers: []EnumMember{},
			}
			t.parseDirectives(directives)
			epkg.Types = append(epkg.Types, t)
		},
		func(funcName, recvName string, recvPointer bool, pos token.Pos, directives []string) {
			epkg.Funcs = append(epkg.Funcs, &Function{
				Context:         ctx,
				pos:             pos,
				PPackage:        ppkg,
				Name:            funcName,
				Params:          []FunctionParam{},
				ReceiverName:    recvName,
				ReceiverPointer: recvPointer,
				FuncId:          ctx.NextMethodId(),
			})
		},
		func(constName string, directives []string) {
			epkg.constantDirectives[constName] = directives
		},
	)

	return epkg
}

func (epkg *Package) resolveEnums() {
	obj2type := make(map[types.Type]*Type)
	for _, etyp := range epkg.Types {
		if !etyp.IsEnum {
			continue
		}
		if basic, ok := etyp.Term.(*ir.Basic); ok && basic.TypeInfo&types.IsInteger != 0 {
			obj2type[etyp.Type()] = etyp
		} else {
			etyp.ThrowAt(etyp, "exported enum must have an underlying type of integer")
		}
	}
	if len(obj2type) == 0 {
		return
	}
	scope := epkg._PPackage.Types.Scope()
	names := scope.Names()
	for _, name := range names {
		objConst, ok := scope.Lookup(name).(*types.Const)
		if !ok {
			continue
		}
		etyp, exists := obj2type[objConst.Type()]
		if !exists {
			continue
		}
		value := objConst.Val()
		if value.Kind() != constant.Int {
			panic("unreachable")
		}

		constName := objConst.Name()
		i64, isInt := constant.Int64Val(value)
		u64, isUint := constant.Uint64Val(value)
		if isUint {
			i64 = int64(u64)
		} else if !isInt {
			panic("unreachable")
		}
		constValue := strconv.AppendInt(nil, i64, 10)

		if directives, ok := epkg.constantDirectives[constName]; ok && len(directives) > 0 {
			constName = directives[0]
		}

		err := etyp.addEnumMember(constName, string(constValue))
		if err != nil {
			etyp.ThrowAt(objConst, "%s", err)
		}
	}

	for _, etyp := range obj2type {
		if len(etyp.EnumMembers) == 0 {
			etyp.ThrowAt(etyp, "exported enum with no relevant constant values defined")
		}
	}
}

func (epkg *Package) Resolve() {
	for _, etyp := range epkg.Types {
		etyp.Resolve()
	}

	epkg.resolveEnums()

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
