package collector

import (
	"github.com/hsfzxjy/dgo/dgo-gen/internal/interfaces"
)

type _IdManager struct {
	nTypes   int64
	nMethods int64
}

func (m *_IdManager) NextTypeId() int64 {
	m.nTypes += 1
	return m.nTypes
}

func (m *_IdManager) NextMethodId() int64 {
	m.nMethods += 1
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
