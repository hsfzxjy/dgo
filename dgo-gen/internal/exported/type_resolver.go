package exported

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type accessKind int

const (
	akInvalid accessKind = iota
	akRoot
	akElem
	akField
	akDeref
	akTypedef
)

type access struct {
	kind      accessKind
	typ       types.Type
	fieldName string
}

type typeResolver struct {
	*NameResolver
	path    []access
	visited map[uri.Uri]struct{}
	terms   []ir.Term
	Result  ir.Term
}

func NewTypeResolver(nameResolver *NameResolver) *typeResolver {
	return &typeResolver{NameResolver: nameResolver}
}

func (r *typeResolver) printPath() string {
	b := strings.Builder{}
	for _, a := range r.path {
		switch a.kind {
		case akRoot:
			b.WriteString(fmt.Sprintf("\n\ttype = %s", a.typ))
		case akElem:
			b.WriteString(fmt.Sprintf(",\n\t\twhose element is\n\ttype = %s", a.typ))
		case akField:
			b.WriteString(fmt.Sprintf(",\n\t\twhose field `.%s` is of\n\ttype = %s", a.fieldName, a.typ))
		case akDeref:
			b.WriteString(fmt.Sprintf(",\n\t\twhich de-references into\n\ttype = %s", a.typ))
		case akTypedef:
			b.WriteString(fmt.Sprintf(",\n\t\twhich is a new type from\n\ttype = %s", a.typ.Underlying().String()))
		}
	}
	return b.String()
}

func (r *typeResolver) throwAt(obj types.Object, args ...any) {
	msg := fmt.Sprintf(args[0].(string), args[1:]...)
	msg = fmt.Sprintf("%s%s", msg, r.printPath())
	r.NameResolver.ThrowAt(obj, msg)
}

func (r *typeResolver) pushPath(kind accessKind, typ types.Type) *access {
	r.path = append(r.path, access{kind, typ, ""})
	return &r.path[len(r.path)-1]
}

func (r *typeResolver) popPath() { r.path = r.path[:len(r.path)-1] }

func (r *typeResolver) push(obj *types.TypeName) (pop func()) {
	uri := uri.UriFor(obj)
	if _, ok := r.visited[uri]; ok {
		r.throwAt(obj, "circular dependency: %s", uri)
		return
	}
	r.visited[uri] = struct{}{}
	return func() { delete(r.visited, uri) }
}

func (r *typeResolver) pushTerm(t ir.Term) {
	if len(r.terms) != 0 {
		r.terms[len(r.terms)-1].AddChild(t)
	}
	r.terms = append(r.terms, t)
}

func (r *typeResolver) popTerm() {
	if len(r.terms) == 1 {
		r.Result = r.terms[0]
	}
	r.terms = r.terms[:len(r.terms)-1]
}

func (r *typeResolver) Do(obj types.Object) {
	if r.visited == nil {
		r.visited = make(map[uri.Uri]struct{})
		r.pushPath(akRoot, obj.Type())
		defer r.popPath()
	}

	var ityp types.Type = obj.Type()
SWITCH:
	switch typ := ityp.(type) {
	case *types.Basic:
		switch typ.Kind() {
		case types.Bool,
			types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
			types.Float32, types.Float64,
			types.String:
			r.pushTerm(ir.NewBasic(typ))
			r.popTerm()
			return
		default:
			r.throwAt(obj, "not exportable type: %s", typ.String())
		}
	case *types.Array:
		ityp = typ.Elem()
		r.pushPath(akElem, ityp)
		defer r.popPath()
		r.pushTerm(ir.NewArray(int(typ.Len())))
		defer r.popTerm()
		goto SWITCH
	case *types.Pointer:
		ityp = typ.Elem()
		r.pushPath(akDeref, ityp)
		defer r.popPath()
		r.pushTerm(ir.NewPtrTo())
		defer r.popTerm()
		goto SWITCH
	case *types.Named:
		if !r.HasTypeName(typ.Obj()) {
			r.throwAt(typ.Obj(), "type not marked as exported: %s", typ)
		}
		pop := r.push(typ.Obj())
		defer pop()
		r.pushPath(akTypedef, typ)
		defer r.popPath()
		r.pushTerm(ir.NewCoerce(typ.Obj().Pkg(), typ.Obj().Name()))
		defer r.popTerm()
		ityp = ityp.Underlying()
		goto SWITCH
	case *types.Struct:
		r.pushTerm(ir.NewStruct())
		defer r.popTerm()
		for i := 0; i < typ.NumFields(); i++ {
			field := typ.Field(i)
			r.pushPath(akField, field.Type()).fieldName = field.Name()
			r.pushTerm(ir.NewField(field.Name()))
			r.Do(field)
			r.popTerm()
			r.popPath()
		}
		return
	default:
		r.throwAt(obj, "not exportable type: %s", typ)
	}
}
