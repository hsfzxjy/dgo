package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type ConfigStruct struct {
	DartProject struct {
		Path        string `ms:"path"`
		GeneratedIn string `ms:"generated-in"`
	} `ms:"dart-project"`
	Packages struct {
		Renames map[string]string `ms:"renames"`
	} `ms:"packages"`
}

func (c *ConfigStruct) Parse() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("dgo.config")
	viper.AddConfigPath(".")
	viper.SetDefault("packages.renames", make(map[string]string))
	exception.Die(viper.ReadInConfig())
	err := viper.UnmarshalExact(&c, func(dc *mapstructure.DecoderConfig) {
		dc.ErrorUnset = true
		dc.ErrorUnused = true
		dc.TagName = "ms"
	})
	exception.Die(err)

	c.DartProject.Path, err = filepath.Abs(c.DartProject.Path)
	exception.Die(err)

	info, err := os.Stat(c.DartProject.Path)
	exception.Die(err)
	if !info.IsDir() {
		exception.Die(fmt.Errorf("%s is not a directory", c.DartProject.Path))
	}
}

var Struct ConfigStruct
