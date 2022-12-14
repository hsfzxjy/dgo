package exported

import (
	"go/token"
	"go/types"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/collector"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"golang.org/x/tools/go/packages"
)

type FunctionParam struct {
	Name string
	Term ir.Term
}

type Function struct {
	*collector.Context `json:"-"`
	pos                token.Pos         `json:"-"`
	PPackage           *packages.Package `json:"-"`
	FuncId             uint32
	Name               string
	Type               *Type  `json:"-"`
	ReceiverName       string `json:"-"`
	ReceiverPointer    bool   `json:"-"`
	Params             []FunctionParam
	Return             ir.Term
	ReturnError        bool
}

func isErrorInterface(t types.Type) bool {
	if iface, ok := t.Underlying().(*types.Interface); ok {
		if iface.NumMethods() != 1 {
			return false
		}
		method := iface.Method(0)
		sig := method.Type().(*types.Signature)
		if method.Name() != "Error" ||
			sig.Params() != nil ||
			sig.Results() == nil {
			return false
		}
		resultType := sig.Results().At(0).Type()
		if t, ok := resultType.(*types.Basic); ok {
			return t.Kind() == types.String
		}
	}
	return false
}

func (efunc *Function) Pos() token.Pos { return efunc.pos }

func (efunc *Function) Resolve() {
	if efunc.Type != nil {
		return
	}
	if typ, ok := efunc.Context.PackageOf(efunc.PPackage.PkgPath).
		Type(efunc.ReceiverName).(*Type); ok {
		efunc.Type = typ
	} else {
		exception.ThrowAt(efunc.PPackage, efunc, "receiver type is not exported")
	}

	var methodSet *types.MethodSet
	if efunc.ReceiverPointer {
		methodSet = types.NewMethodSet(types.NewPointer(efunc.Type.Type()))
	} else {
		methodSet = types.NewMethodSet(efunc.Type.Type())
	}

	f := methodSet.Lookup(efunc.Type.Pkg(), efunc.Name)
	sig := f.Type().(*types.Signature)
	if sig.Variadic() {
		efunc.ThrowAt(f.Obj(), "variadic function is not exportable")
	}
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		paramName := param.Name()
		if len(paramName) == 0 {
			efunc.ThrowAt(param, "function with unnamed parameter is not exportable")
		}
		paramTerm := efunc.SolveType(param)
		efunc.Params = append(efunc.Params, FunctionParam{paramName, paramTerm})
	}
	results := sig.Results()
	if results != nil {
		switch results.Len() {
		case 1:
			if isErrorInterface(results.At(0).Type()) {
				efunc.ReturnError = true
			} else {
				efunc.Return = efunc.SolveType(results.At(0))
			}
		case 2:
			efunc.Return = efunc.SolveType(results.At(0))
			if !isErrorInterface(results.At(1).Type()) {
				goto BAD_RETURN_SIG
			}
			efunc.ReturnError = true
		default:
			goto BAD_RETURN_SIG
		}
	}

	efunc.Type.Methods = append(efunc.Type.Methods, TypeMethod{efunc.Name, efunc})
	return

BAD_RETURN_SIG:
	efunc.ThrowAt(f.Obj(), `exported function should return
 1) nothing OR
 2) type T OR
 3) type error OR
 4) type (T, error)
where T is an exportable type.`)
}
