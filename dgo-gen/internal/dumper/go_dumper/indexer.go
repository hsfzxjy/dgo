package go_dumper

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
)

type rep struct {
	N, Size, Len int
	Index        *Statement
}

type indexer struct {
	n    int
	reps []rep
}

func (i *indexer) Next() (ret *Statement) {
	if len(i.reps) == 0 {
		ret = Lit(i.n)
		i.n++
	} else {
		nreps := len(i.reps)
		last := &i.reps[nreps-1]
		ret = Parens(last.Index.Clone().Op("+").Lit(last.N))
		last.N++
		for last.N == last.Size {
			NSkip := last.Size * last.Len
			i.reps = i.reps[:nreps-1]
			nreps--
			if nreps == 0 {
				i.n += NSkip
				break
			} else {
				last = &i.reps[nreps-1]
				last.N += NSkip
			}
		}
	}
	return
}

func (i *indexer) BeginRep(size, l int) *Statement {
	looper := Id(fmt.Sprintf("i%d", len(i.reps)))
	if len(i.reps) == 0 {
		i.reps = append(i.reps, rep{
			N:     0,
			Size:  size,
			Len:   l,
			Index: Lit(i.n).Op("+").Add(looper).Op("*").Lit(size),
		})
	} else {
		last := i.reps[len(i.reps)-1]
		i.reps = append(i.reps, rep{
			N:     0,
			Size:  size,
			Len:   l,
			Index: last.Index.Clone().Op("+").Lit(last.N).Op("+").Add(looper).Op("*").Lit(size),
		})
	}
	return looper
}
