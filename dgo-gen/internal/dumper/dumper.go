package dumper

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
)

type Dumper interface {
	AddType(*exported.Type)
	Save()
}
