package collector

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"golang.org/x/tools/go/packages"
)

func (ctx *Context) Collect(
	ppkg *packages.Package,
	onType func(typeName string),
	onFunction func(funcName, recvName string, recvPointer bool),
) {
	isExported := func(doc *ast.CommentGroup) bool {
		return doc != nil && len(doc.List) > 0 && strings.HasPrefix(doc.List[0].Text, "//dgo:export")
	}

	for _, file := range ppkg.Syntax {
	NEXT_DECL:
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok != token.TYPE {
					continue NEXT_DECL
				}
				declExported := isExported(decl.Doc)
				for _, spec := range decl.Specs {
					spec := spec.(*ast.TypeSpec)
					if declExported || isExported(spec.Doc) {
						if spec.Assign != token.NoPos {
							exception.ThrowAt(ppkg, spec, "type alias is not exportable")
						}
						onType(spec.Name.Name)
					}
				}
			case *ast.FuncDecl:
				if !isExported(decl.Doc) {
					continue NEXT_DECL
				}
				name := decl.Name.Name
				if decl.Recv == nil {
					exception.ThrowAt(ppkg, decl, "non-method function is not exportable")
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

				onFunction(name, recvName, recvPtr)
			}
		}
	}
}
