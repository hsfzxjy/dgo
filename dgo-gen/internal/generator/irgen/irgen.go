package irgen

import (
	"encoding/json"
	"os"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type Entry struct {
	ir.Term
	TypeId  int64
	Methods []exported.TypeMethod
}

type Generator struct {
	m map[uri.Uri]Entry
}

func (d *Generator) AddType(etype *exported.Type) {
	if d.m == nil {
		d.m = make(map[uri.Uri]Entry)
	}
	uri := etype.Uri()
	d.m[uri] = Entry{etype.Term, etype.TypeId, etype.Methods}
}

func (d *Generator) Save() {
	out, _ := json.MarshalIndent(d.m, "", "  ")
	os.WriteFile("tests/gen_tests_dart/ir.json", out, 0o644)
}
