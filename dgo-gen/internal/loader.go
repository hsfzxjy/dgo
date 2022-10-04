package internal

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/packages"
)

func printErrors(pkgs []*packages.Package) int {
	n := 0
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		for _, err := range p.Errors {
			isCgoError := strings.Contains(err.Msg, `import C`)
			suffix := ""
			if isCgoError {
				suffix = " (IGNORED)"
			} else {
				n++
			}
			fmt.Fprintf(os.Stderr, "%s%s\n", err, suffix)
		}
	})
	return n
}

func LoadPackages(pattern string) []*packages.Package {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule,
	}, pattern)
	Die(err)
	if printErrors(pkgs) > 0 {
		Exit()
	}
	return pkgs
}
