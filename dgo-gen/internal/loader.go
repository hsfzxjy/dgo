package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"golang.org/x/tools/go/packages"
)

func printErrors(pkgs []*packages.Package) int {
	n := 0
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		for _, err := range p.Errors {
			isCgoError := strings.Contains(err.Msg, `import C`)
			if isCgoError {
				continue
			} else {
				n++
			}
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	})
	return n
}

func LoadPackages(pattern string) []*packages.Package {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule,
	}, pattern)
	exception.Die(err)
	if printErrors(pkgs) > 0 {
		// TODO
	}
	return pkgs
}
