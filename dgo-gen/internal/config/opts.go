package config

type OptsStruct struct {
	GoDir struct {
		Value string `positional-arg-name:"go-dir" required:"1"`
	} `positional-args:"yes"`
	ConfigFile string `short:"c" long:"config" default:"dgo.config.yaml" description:"Path to dgo generator config file"`
}

var Opts OptsStruct
