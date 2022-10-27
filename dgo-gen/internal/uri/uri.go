package uri

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

type Uri string

func UriFor(pkg *packages.Package, name string) Uri {
	modPath := pkg.Module.Path
	pkgPath := strings.TrimPrefix(pkg.PkgPath, modPath)
	pkgPath = strings.TrimLeft(pkgPath, "/")
	return Uri(strings.Join([]string{modPath, pkgPath, name}, "#"))
}

func (uri Uri) IsWithPackagePath(pkgPath string) bool {
	pos := strings.LastIndexByte(string(uri), '#')
	return uri[:pos] == Uri(pkgPath)
}

func (uri Uri) IsSamePackage(other Uri) bool {
	pos2 := strings.LastIndexByte(string(uri), '#')
	return uri.IsWithPackagePath(string(other[:pos2]))
}
