package exported

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"golang.org/x/tools/go/packages"
)

type Function struct {
	*NameResolver
	OriPkg          *packages.Package
	Name            string
	ReceiverName    string
	ReceiverPointer bool
}

func isExported(doc *ast.CommentGroup) bool {
	return doc != nil && len(doc.List) > 0 && strings.HasPrefix(doc.List[0].Text, "//dgo:export")
}

type Package struct {
	OriPkg *packages.Package
	Types  map[string]*Type
	Funcs  []*Function
	IsDep  bool
}

func NewPackage(pkg *packages.Package, nameResolver *NameResolver, isDep bool) *Package {
	epkg := &Package{
		OriPkg: pkg,
		Types:  make(map[string]*Type),
		IsDep:  isDep,
	}

	for _, file := range pkg.Syntax {
		parseTypeSpec := func(spec *ast.TypeSpec) {
			name := spec.Name.Name
			epkg.Types[name] = &Type{
				NameResolver: nameResolver,
				Object:       pkg.Types.Scope().Lookup(name),
				OriPkg:       pkg,
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
								exception.ThrowAt(pkg, spec, "type alias is not exportable")
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
					exception.ThrowAt(pkg, decl, "non-method function is not exportable")
					return nil
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

				epkg.Funcs = append(epkg.Funcs, &Function{
					NameResolver:    nameResolver,
					OriPkg:          pkg,
					Name:            name,
					ReceiverName:    recvName,
					ReceiverPointer: recvPtr,
				})
			}
		}
	}
	return epkg
}

func (ex *Package) Resolve() {
	for _, etyp := range ex.Types {
		etyp.Resolve()
	}
}
