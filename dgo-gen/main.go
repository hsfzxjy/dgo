package main

import (
	"os"

	"github.com/hsfzxjy/dgo/dgo-gen/internal"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/config"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/jessevdk/go-flags"

	// retain spew as dependency, which is somehow useful for debugging
	_ "github.com/davecgh/go-spew/spew"
)

func main() {
	parser := flags.NewParser(&config.Opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := parser.Parse()
	exception.Die(err)
	exception.Die(os.Chdir(config.Opts.GoDir.Value))

	config.Struct.Parse()

	exports := internal.NewExports(".")
	exports.Dump()
}
