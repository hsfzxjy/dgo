package interfaces

import "golang.org/x/tools/go/packages"

type Type interface{}

type Package interface {
	PPackage() *packages.Package
	HasType(name string) bool
	Type(name string) Type
}

type Storage interface {
	ParsePackages(pkgPath string, isDep bool)
	Package(name string) Package
}
