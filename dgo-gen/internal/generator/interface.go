package generator

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
)

type Interface interface {
	AddType(*exported.Type)
	Save()
}
