package irgen

import (
	"encoding/json"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/config"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type Entry struct {
	ir.Term
	TypeId  int64
	Methods []exported.TypeMethod
}

type Payload struct {
	Config      *config.ConfigStruct
	Definitions map[uri.Uri]Entry
}

type Generator struct {
	payload Payload
}

func (d *Generator) AddType(etype *exported.Type) {
	if d.payload.Definitions == nil {
		d.payload.Definitions = make(map[uri.Uri]Entry)
	}
	uri := etype.Uri()
	d.payload.Definitions[uri] = Entry{etype.Term, etype.TypeId, etype.Methods}
}

func (d *Generator) Save() {
	d.payload.Config = &config.Struct
	var err error
	MarshaledPayload, err = json.MarshalIndent(d.payload, "", "  ")
	exception.Die(err)
}

var MarshaledPayload []byte
