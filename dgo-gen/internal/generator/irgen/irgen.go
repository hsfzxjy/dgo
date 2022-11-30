package irgen

import (
	"encoding/json"
	"os"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/config"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type Payload struct {
	Config      *config.ConfigStruct
	Definitions map[uri.Uri]*exported.Type
}

type Generator struct {
	payload Payload
}

func (d *Generator) AddType(etype *exported.Type) {
	if d.payload.Definitions == nil {
		d.payload.Definitions = make(map[uri.Uri]*exported.Type)
	}
	uri := etype.Uri()
	d.payload.Definitions[uri] = etype
}

func (d *Generator) Save() {
	d.payload.Config = &config.Struct
	var err error
	MarshaledPayload, err = json.MarshalIndent(d.payload, "", "  ")
	exception.Die(err)
	if config.Opts.IrFile != "" {
		file, err := os.Create(config.Opts.IrFile)
		exception.Die(err)
		defer file.Close()
		file.Write(MarshaledPayload)
	}
}

var MarshaledPayload []byte
