package gogen

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
)

type looper struct {
	n int
}

func (i *looper) BeginRep() *Statement {
	i.n++
	return Id(fmt.Sprintf("i%d", i.n))
}

func (i *looper) EndRep() {
	i.n--
}
