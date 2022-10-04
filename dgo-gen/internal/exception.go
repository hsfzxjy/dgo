package internal

import (
	"fmt"
	"go/token"
	"os"

	"golang.org/x/tools/go/packages"
)

func Throw(args ...any) {
	text := fmt.Sprintf(args[0].(string), args[1:]...)
	fmt.Fprintf(os.Stderr, "error: %s\n", text)
	os.Exit(1)
}

type hasPos interface {
	Pos() token.Pos
}

func throwAt(pkg *packages.Package, obj hasPos, args ...any) {
	text := fmt.Sprintf(args[0].(string), args[1:]...)
	Throw("%s: %s", pkg.Fset.Position(obj.Pos()), text)
}

func Exit() { os.Exit(1) }

func Die(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "oops: %v\n", err)
		Exit()
	}
}
