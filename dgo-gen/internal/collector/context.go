package collector

import (
	"errors"
	"math"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/interfaces"
)

type _IdManager struct {
	nTypes   uint16
	nMethods uint32
}

func (m *_IdManager) NextTypeId() uint16 {
	m.nTypes += 1
	if m.nTypes == math.MaxUint16 {
		exception.Die(errors.New("too many exported types"))
	}
	return m.nTypes
}

func (m *_IdManager) NextMethodId() uint32 {
	m.nMethods += 1
	if m.nMethods == math.MaxUint32 {
		exception.Die(errors.New("too many exported functions"))
	}
	return m.nMethods
}

type Context struct {
	*NameSolver
	_IdManager
}

func NewContext(m interfaces.Storage) *Context {
	return &Context{
		NameSolver: NewNameSolver(m),
	}
}
