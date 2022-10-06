package uri

import (
	"fmt"
	"go/types"
	"strings"
)

type Uri string

func UriFor(obj *types.TypeName) Uri {
	return NewUri(obj.Pkg().Path(), obj.Name())
}

func NewUri(pkgPath, name string) Uri {
	return Uri(fmt.Sprintf("%s#%s", pkgPath, name))
}

func (uri Uri) IsWithPackagePath(pkgPath string) bool {
	pos := strings.IndexByte(string(uri), '#')
	return uri[:pos] == Uri(pkgPath)
}

func (uri Uri) IsSamePackage(other Uri) bool {
	pos2 := strings.IndexByte(string(uri), '#')
	return uri.IsWithPackagePath(string(other[:pos2]))
}
