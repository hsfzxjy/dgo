package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"golang.org/x/tools/go/packages"
)

type uid string

func uidFor(obj *types.TypeName) uid {
	return uid(fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name()))
}

type ExportedType struct {
	types.Object
	*Exports
	pkg       *packages.Package
	resolved  bool
	irTerm ir.Term
}

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

type resolver struct {
	path    []access
	visited map[uid]struct{}
	*ExportedType
	terms []ir.Term
}

func (r *resolver) printPath() string {
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

func (r *resolver) throwAt(obj types.Object, args ...any) {
	msg := fmt.Sprintf(args[0].(string), args[1:]...)
	msg = fmt.Sprintf("%s%s", msg, r.printPath())
	r.Exports.throwAt(obj, msg)
}

func (r *resolver) pushPath(kind accessKind, typ types.Type) *access {
	r.path = append(r.path, access{kind, typ, ""})
	return &r.path[len(r.path)-1]
}

func (r *resolver) popPath() { r.path = r.path[:len(r.path)-1] }

func (r *resolver) push(obj *types.TypeName) (pop func()) {
	uid := uidFor(obj)
	if _, ok := r.visited[uid]; ok {
		r.throwAt(obj, "circular dependency: %s", uid)
		return
	}
	r.visited[uid] = struct{}{}
	return func() { delete(r.visited, uid) }
}

func (r *resolver) pushTerm(t ir.Term) {
	if len(r.terms) != 0 {
		r.terms[len(r.terms)-1].AddChild(t)
	}
	r.terms = append(r.terms, t)
}

func (r *resolver) popTerm() {
	if len(r.terms) == 1 {
		r.irTerm = r.terms[0]
	}
	r.terms = r.terms[:len(r.terms)-1]
}

func (r *resolver) do(obj types.Object) {
	if r.visited == nil {
		r.visited = make(map[uid]struct{})
		r.ExportedType = obj.(*ExportedType)

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
			r.pushTerm(&ir.Basic{Type: typ})
			r.popTerm()
			return
		default:
			r.throwAt(obj, "not exportable type: %s", typ.String())
		}
	case *types.Array:
		ityp = typ.Elem()
		r.pushPath(akElem, ityp)
		defer r.popPath()
		r.pushTerm(&ir.Array{Len: typ.Len()})
		defer r.popTerm()
		goto SWITCH
	case *types.Pointer:
		ityp = typ.Elem()
		r.pushPath(akDeref, ityp)
		defer r.popPath()
		r.pushTerm(&ir.PtrTo{})
		defer r.popTerm()
		goto SWITCH
	case *types.Named:
		if !r.hasTypeName(typ.Obj()) {
			r.throwAt(typ.Obj(), "type not marked as exported: %s", typ)
		}
		pop := r.push(typ.Obj())
		defer pop()
		r.pushPath(akTypedef, typ)
		defer r.popPath()
		r.pushTerm(&ir.Coerce{Ident: ir.Ident{Pkg: typ.Obj().Pkg(), Name: typ.Obj().Name()}})
		defer r.popTerm()
		ityp = ityp.Underlying()
		goto SWITCH
	case *types.Struct:
		r.pushTerm(&ir.Struct{})
		defer r.popTerm()
		for i := 0; i < typ.NumFields(); i++ {
			field := typ.Field(i)
			r.pushPath(akField, field.Type()).fieldName = field.Name()
			r.pushTerm(&ir.Field{Name: field.Name()})
			r.do(field)
			r.popTerm()
			r.popPath()
		}
		return
	default:
		r.throwAt(obj, "not exportable type: %s", typ)
	}
}

func (etyp *ExportedType) resolve() {
	if etyp.resolved {
		return
	}
	(&resolver{}).do(etyp)
}

func isExported(doc *ast.CommentGroup) bool {
	return doc != nil && len(doc.List) > 0 && strings.HasPrefix(doc.List[0].Text, "//dgo:export")
}

func parsePackage(pkg *packages.Package, epkg *ExportedPkg) {
	for _, file := range pkg.Syntax {
		// ast.Print(pkg.Fset, file)
		parseTypeSpec := func(spec *ast.TypeSpec) {
			name := spec.Name.Name
			epkg.Types[name] = &ExportedType{
				Exports: epkg.Exports,
				Object:  pkg.Types.Scope().Lookup(name),
				pkg:     pkg,
			}
		}
	NEXT_DECL:
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok != token.TYPE {
					continue NEXT_DECL
				}
				if decl.Lparen == token.NoPos {
					if isExported(decl.Doc) {
						parseTypeSpec(decl.Specs[0].(*ast.TypeSpec))
					}
				} else {
					for _, spec := range decl.Specs {
						spec := spec.(*ast.TypeSpec)
						if isExported(spec.Doc) {
							if spec.Assign != token.NoPos {
								throwAt(pkg, spec, "type alias is not exportable")
							}
							parseTypeSpec(spec)
						}
					}
				}
			case *ast.FuncDecl:
				if !isExported(decl.Doc) {
					continue NEXT_DECL
				}
				name := decl.Name.Name
				if decl.Recv == nil {
					throwAt(pkg, decl, "non-method function is not exportable")
					return
				}

				var recvName string
				var recvPtr bool

				switch recv := decl.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					recvPtr = true
					recvName = recv.X.(*ast.Ident).Name
				case *ast.Ident:
					recvPtr = false
					recvName = recv.Name
				}

				epkg.Funcs = append(epkg.Funcs, &ExportedFunc{
					Exports:         epkg.Exports,
					Pkg:             pkg,
					Name:            name,
					ReceiverName:    recvName,
					ReceiverPointer: recvPtr,
				})
			}
		}
	}
}

type ExportedFunc struct {
	Exports         *Exports
	Pkg             *packages.Package
	Name            string
	ReceiverName    string
	ReceiverPointer bool
}

type ExportedPkg struct {
	Pkg     *packages.Package
	Exports *Exports
	Types   map[string]*ExportedType
	Funcs   []*ExportedFunc
	IsDep   bool
}

type Exports struct {
	Pkgs map[string]*ExportedPkg
}

func (ex *Exports) parsePackages(pkgs []*packages.Package, isDep bool) {
	for _, pkg := range pkgs {
		epkg := &ExportedPkg{
			Pkg:     pkg,
			Exports: ex,
			Types:   make(map[string]*ExportedType),
			IsDep:   isDep,
		}
		ex.Pkgs[pkg.PkgPath] = epkg
		parsePackage(pkg, epkg)
	}
}

func ParsePackages(pkgs []*packages.Package) *Exports {
	var exports = &Exports{
		Pkgs: make(map[string]*ExportedPkg),
	}
	exports.parsePackages(pkgs, false)
	exports.resolve()
	return exports
}

func (ex *ExportedPkg) resolve() {
	for _, etyp := range ex.Types {
		etyp.resolve()
		// spew.Dump(etyp.irTerm)
	}
}

func (ex *Exports) resolve() {
	for _, epkg := range ex.Pkgs {
		epkg.resolve()
	}
}

func (ex *Exports) hasTypeName(obj *types.TypeName) bool {
	pkgPath := obj.Pkg().Path()
TRY_LOAD_PKG:
	if epkg, pkgExists := ex.Pkgs[pkgPath]; pkgExists {
		_, exists := epkg.Types[obj.Name()]
		return exists
	} else {
		ex.parsePackages(LoadPackages(pkgPath), true)
		goto TRY_LOAD_PKG
	}
}

func (ex *Exports) throwAt(obj types.Object, args ...any) {
	pkgPath := obj.Pkg().Path()
	pkg := ex.Pkgs[pkgPath].Pkg
	throwAt(pkg, obj, args...)
}
