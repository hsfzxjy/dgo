package main

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal"
	"github.com/jessevdk/go-flags"
)

type Config struct {
	PackageName struct {
		Value string `positional-arg-name:"package" required:"1"`
	} `positional-args:"yes"`
}

var config Config

func main() {
	parser := flags.NewParser(&config, flags.HelpFlag|flags.PassDoubleDash)
	_, err := parser.Parse()
	internal.Die(err)
	pkgs := internal.LoadPackages(config.PackageName.Value)
	ex := internal.ParsePackages(pkgs)
	ex.Dump()
}
