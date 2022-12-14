package collector

import (
	"bytes"
	"fmt"
	"go/types"
	"reflect"
	"regexp"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type TypeSolveConfig struct {
	IsPinnable bool
}

func (ctx *Context) SolveType(obj types.Object) ir.Term {
	return ctx.SolveTypeEx(obj, TypeSolveConfig{}).Result
}

func (ctx *Context) SolveTypeEx(obj types.Object, config TypeSolveConfig) *typeSolver {
	solver := &typeSolver{Context: ctx, config: config}
	solver.Do(obj, nil)
	term := solver.Result
	ir.FillAllSize(term)
	return solver
}

type termLayerKind int

const (
	tlkInvalid termLayerKind = iota
	tlkElem
	tlkField
	tlkDeref
	tlkTypedef
	tlkOther
)

type termLayer struct {
	kind      termLayerKind
	typ       types.Type
	fieldName string
	term      ir.Term
}

type typeSolver struct {
	*Context
	config  TypeSolveConfig
	layers  []termLayer
	visited map[uri.Uri]struct{}
	Result  ir.Term
}

func (r *typeSolver) Do(obj types.Object, ityp types.Type) {
	if r.visited == nil {
		r.visited = make(map[uri.Uri]struct{})
	}

	if ityp == nil {
		ityp = obj.Type()
	}
SWITCH:
	switch typ := ityp.(type) {
	case *types.Basic:
		switch typ.Kind() {
		case types.Bool,
			types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
			types.Float32, types.Float64,
			types.String:
			r.push(tlkOther, typ, ir.NewBasic(typ))
			r.pop()
			return
		default:
			r.throwAt(obj, "not exportable type: %s", typ.String())
		}
	case *types.Slice:
		r.push(tlkElem, typ, ir.NewSlice())
		defer r.pop()
		ityp = typ.Elem()
		goto SWITCH
	case *types.Map:
		irMap := ir.NewMap()
		r.push(tlkOther, typ, irMap)
		defer r.pop()
		r.Do(obj, typ.Key())
		r.Do(obj, typ.Elem())

		irKey := irMap.Key
	SWITCH_KEY:
		switch key := irKey.(type) {
		case *ir.Basic:
		case *ir.Coerce:
			irKey = key.Elem
			goto SWITCH_KEY
		default:
			r.throwAt(obj, "map type with key %s is not exportable", typ.Key().String())
		}
	case *types.Array:
		r.push(tlkElem, typ, ir.NewArray(int(typ.Len())))
		defer r.pop()
		ityp = typ.Elem()
		goto SWITCH
	case *types.Pointer:
		r.push(tlkDeref, typ, ir.NewPtrTo())
		defer r.pop()
		ityp = typ.Elem()
		goto SWITCH
	case *types.Named:
		if isOptional(typ) {
			ityp = ityp.Underlying().(*types.Struct).Field(0).Type()
			r.push(tlkField, ityp, ir.NewOptional()).fieldName = "Value"
			defer r.pop()

			goto SWITCH
		}
		if isPinToken(typ) {
			ityp = typ.TypeArgs().At(0)
			r.push(tlkField, ityp, ir.NewPinToken()).fieldName = "Underlying"
			defer r.pop()

			goto SWITCH
		}
		if !r.HasTypeName(typ.Obj()) {
			r.throwAt(obj, "type not marked as exported: %s", typ)
		}
		r.push(tlkTypedef, ityp, ir.NewCoerce(typ.Obj(), r.UriFor(typ.Obj())))
		defer r.pop()
		ityp = ityp.Underlying()
		goto SWITCH
	case *types.Chan:
		if !r.isChanAllowedHere() {
			r.throwAt(obj, "chan type is not exportable")
		}
		if typ.Dir() == types.SendOnly {
			r.throwAt(obj, "send-only chan type is not exportable")
		}
		r.push(tlkElem, ityp, ir.NewChan())
		defer r.pop()
		ityp = typ.Elem()
		goto SWITCH
	case *types.Struct:
		if !r.isTypeNamed() {
			r.throwAt(obj, "anonymous type is not exportable")
		}
		irStruct := ir.NewStruct()
		r.push(tlkOther, typ, irStruct)
		defer r.pop()

		var i = 0
		if r.config.IsPinnable && r.isRootLevelStruct() {
			if typ.NumFields() == 0 {
				goto BAD_PINNABLE
			}
			i = 1
			field := typ.Field(0)
			if !isPinMeta(field.Type()) || !field.Embedded() {
				goto BAD_PINNABLE
			}
		}
		goto NEXT_FIELD
	BAD_PINNABLE:
		r.throwAt(obj, "pinnable struct should have an embedded pin.Meta as the first field")

	NEXT_FIELD:
		for ; i < typ.NumFields(); i++ {
			field := typ.Field(i)

			if i == 0 && isPinMeta(field.Type()) {
				continue NEXT_FIELD
			}

			var directives *ir.FieldDirectives
			{
				tag := reflect.StructTag(typ.Tag(i))
				spec, _ := tag.Lookup("dgo")
				directives = ir.ParseFieldDirectives(spec)
			}

			switch {
			case !directives.SendToDart:
				continue NEXT_FIELD
			}

			irField := ir.NewField(field.Name(), directives)
			if irStruct.IsFieldNameConflicted(irField.DartName()) {
				r.throwAt(field, "conflicted field name: %s", irField.DartName())
			}

			r.push(tlkField, field.Type(), irField).
				fieldName = field.Name()
			r.Do(field, nil)
			r.pop()

			nFields := len(irStruct.Fields)
			if c, ok := irField.Term.(*ir.Chan); ok {
				irStruct.Fields = irStruct.Fields[:nFields-1]
				irStruct.Chans = append(irStruct.Chans, irField)
				c.Chid = uint8(irStruct.Nchans)
				irStruct.Nchans++
			}
		}
		return
	default:
		r.throwAt(obj, "not exportable type: %s", typ)
	}
}

func (r *typeSolver) printPath() string {
	if len(r.layers) == 0 {
		return ""
	}
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("\n\ttype = %s", r.layers[0].typ))
	for _, a := range r.layers {
		switch a.kind {
		case tlkElem:
			b.WriteString(fmt.Sprintf(",\n\t\twhose element is\n\ttype = %s", a.typ.(interface{ Elem() types.Type }).Elem()))
		case tlkField:
			b.WriteString(fmt.Sprintf(",\n\t\twhose field `.%s` is of\n\ttype = %s", a.fieldName, a.typ))
		case tlkDeref:
			b.WriteString(fmt.Sprintf(",\n\t\twhich de-references into\n\ttype = %s", a.typ.(*types.Pointer).Elem()))
		case tlkTypedef:
			b.WriteString(fmt.Sprintf(",\n\t\twhich is a new type from\n\ttype = %s", a.typ.Underlying().String()))
		}
	}
	return b.String()
}

func (r *typeSolver) throwAt(obj types.Object, args ...any) {
	msg := fmt.Sprintf(args[0].(string), args[1:]...)
	msg = fmt.Sprintf("%s%s", msg, r.printPath())
	r.NameSolver.ThrowAt(obj, msg)
}

func (r *typeSolver) push(kind termLayerKind, typ types.Type, term ir.Term) *termLayer {
	length := len(r.layers)
	if length > 0 {
		if length == 1 {
			lasttermLayer := &r.layers[0]
			if lastTerm, ok := lasttermLayer.term.(*ir.Coerce); ok {
				term.(ir.HasIdent).SetIdent(lastTerm.Ident)
				lasttermLayer.term = term
				goto SKIP_ADD_CHILD
			}
		}
		r.layers[length-1].term.AddChild(term)
	SKIP_ADD_CHILD:
	}
	r.layers = append(r.layers, termLayer{kind: kind, typ: typ, term: term})
	lasttermLayer := &r.layers[length]

	if typ, ok := typ.(*types.Named); ok && kind == tlkTypedef {
		uri := r.UriFor(typ.Obj())
		if _, ok := r.visited[uri]; ok {
			r.throwAt(typ.Obj(), "circular dependency: %s", uri)
			return nil
		}
		r.visited[uri] = struct{}{}
	}

	return lasttermLayer
}

func (r *typeSolver) pop() {
	length := len(r.layers)
	lasttermLayer := r.layers[length-1]
	r.layers = r.layers[:length-1]
	if typ, ok := lasttermLayer.typ.(*types.Named); ok && lasttermLayer.kind == tlkTypedef {
		uri := r.UriFor(typ.Obj())
		delete(r.visited, uri)
	}
	if length == 1 {
		r.Result = lasttermLayer.term
	}
}

func (r *typeSolver) matchPattern(pat *regexp.Regexp) bool {
	buf := bytes.Buffer{}
	var prev ir.Term
	for _, l := range r.layers {
		t := l.term
		if t == prev {
			continue
		}
		buf.WriteByte(t.GetHeader().Abbr)
		prev = t
	}
	return pat.Match(buf.Bytes())
}

func (r *typeSolver) isChanAllowedHere() bool {
	if r.matchPattern(regexp.MustCompile("^(.*[^c]|)Sf$")) {
		return r.config.IsPinnable
	}
	return r.matchPattern(regexp.MustCompile("^.*cSf$"))
}

func (r *typeSolver) isRootLevelStruct() bool {
	return r.matchPattern(regexp.MustCompile("^[^S]*S$"))
}

func (r *typeSolver) isTypeNamed() bool {
	var i int
	var term ir.Term
	for i = len(r.layers) - 1; i >= 0; i-- {
		term = r.layers[i].term
		if _, ok := term.(*ir.PtrTo); !ok {
			if _, ok = term.(*ir.Coerce); ok {
				return true
			}
		}
	}
	return false
}

func isDgoType(typ types.Type, subPath, name string) bool {
	if named, ok := typ.(*types.Named); ok {
		obj := named.Obj()
		return obj.Pkg() != nil &&
			obj.Pkg().Path() == "github.com/hsfzxjy/dgo/go"+subPath &&
			obj.Name() == name
	}
	return false
}

func isPinMeta(typ types.Type) bool {
	return isDgoType(typ, "/pin", "Meta")
}

func isOptional(typ types.Type) bool {
	return isDgoType(typ, "", "Optional")
}

func isPinToken(typ types.Type) bool {
	return isDgoType(typ, "/pin", "Token")
}
