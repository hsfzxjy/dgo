package collector

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/utils"
	"golang.org/x/tools/go/packages"
)

var reDgoExport = regexp.MustCompile(`^//\s*dgo:export(.*)$`)

func parseDgoExported(doc *ast.CommentGroup) (directives []string, isExported bool) {
	isExported = false
	if doc == nil || len(doc.List) == 0 {
		return
	}

	matches := reDgoExport.FindStringSubmatch(doc.List[0].Text)
	if len(matches) == 0 {
		return
	}
	isExported = true
	directives = utils.ParseDirectives(strings.TrimSpace(matches[1]))
	return
}

func (ctx *Context) Collect(
	ppkg *packages.Package,
	onType func(typeName string, directives []string),
	onFunction func(funcName, recvName string, recvPointer bool, directives []string),
	onConst func(constName string, directives []string),
) {
	for _, file := range ppkg.Syntax {
	NEXT_DECL:
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				var isExported bool
				var directives []string
				resolveDoc := func(docs ...*ast.CommentGroup) *ast.CommentGroup {
					for _, doc := range docs {
						if doc != nil {
							return doc
						}
					}
					return decl.Doc
				}
				switch decl.Tok {
				case token.TYPE:
					for _, spec := range decl.Specs {
						spec := spec.(*ast.TypeSpec)
						directives, isExported = parseDgoExported(resolveDoc(spec.Doc))
						if isExported {
							if spec.Assign != token.NoPos {
								exception.ThrowAt(ppkg, spec, "type alias is not exportable")
							}
							onType(spec.Name.Name, directives)
						}
					}
				case token.CONST:
					for _, spec := range decl.Specs {
						spec := spec.(*ast.ValueSpec)
						directives, isExported = parseDgoExported(resolveDoc(spec.Comment, spec.Doc))
						if isExported {
							onConst(spec.Names[0].Name, directives)
						}
					}
				}
			case *ast.FuncDecl:
				directives, isExported := parseDgoExported(decl.Doc)
				if !isExported {
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

				onFunction(name, recvName, recvPtr, directives)
			}
		}
	}
}
