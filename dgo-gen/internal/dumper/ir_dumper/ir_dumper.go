package ir_dumper

import (
	"encoding/json"
	"reflect"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type TermWithOpName struct {
	ir.Term
	Op string
}

func newTermWithOpName(term ir.Term) *TermWithOpName {
	opName := reflect.TypeOf(term).Elem().Name()
	return &TermWithOpName{term, opName}
}

type IrDumper struct {
	m map[uri.Uri]TermWithOpName
}

func (d *IrDumper) AddType(etype *exported.Type) {
	if d.m == nil {
		d.m = make(map[uri.Uri]TermWithOpName)
	}
	uri := uri.NewUri(etype.OriPkg.PkgPath, etype.Name())
	d.m[uri] = *newTermWithOpName(etype.IrTerm)
}

func (d *IrDumper) Save() {
	out, _ := json.MarshalIndent(d.m, "", "  ")
	println(string(out))
}
