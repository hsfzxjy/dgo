package internal

import (
	"fmt"
	"go/types"
	"path"

	. "github.com/dave/jennifer/jen"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
)

const dgoMod = "github.com/hsfzxjy/dgo/go"

func dumpTypeName(ex *ExportedType, t ir.Term) *Statement {
	s := Statement(nil)
LOOP:
	for {
		switch tt := t.(type) {
		case *ir.PtrTo:
			s.Op("*")
			t = tt.Elem
		case *ir.Array:
			s.Index(Lit(tt.Len))
			t = tt.Elem
		case *ir.Coerce:
			if ex.pkg.PkgPath == tt.Pkg.Path() {
				s.Id(tt.Name)
			} else {
				s.Qual(tt.Pkg.Path(), tt.Name)
			}
			t = tt.Elem
		case *ir.Basic:
			s.Id(tt.TypeName())
			break LOOP
		case *ir.Struct:
			break LOOP
		default:
			panic("unreachable")
		}
	}
	return &s
}

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

func (ex *ExportedType) dumpLoadInplaceInit(it ir.Term, g *Group, indexer *indexer) {
SWITCH:
	switch t := it.(type) {
	case *ir.Struct:
		for _, f := range t.Fields {
			ex.dumpLoadInplaceInit(f, g, indexer)
		}
	case *ir.Field:
		g.BlockFunc(func(g *Group) {
			g.Id("o").Op(":=").Op("&").Id("o").Dot(t.Name)
			ex.dumpLoadInplaceInit(t.Term, g, indexer)
		})
	case *ir.PtrTo:
		g.Op("*").Id("o").Op("=").New(dumpTypeName(ex, t))
		it = t.Elem
		goto SWITCH
	case *ir.Coerce:
		if el, ok := t.Elem.(*ir.Basic); ok {
			g.BlockFunc(func(g *Group) {
				g.Id("o").Op(":=").Parens(Op("*").Id(el.TypeName())).Parens(Id("o"))
				ex.dumpLoadInplaceInit(t.Elem, g, indexer)
			})
		} else {
			it = t.Elem
			goto SWITCH
		}
	case *ir.Basic:
		info := t.Type.Info()
		switch {
		case info&types.IsBoolean != 0:
			g.Add(codeLoadBasic(Id("arr"), indexer.Next(), Qual("C", "bool"), Id("o"), Id("bool")))
		case info&types.IsInteger != 0:
			g.Add(codeLoadBasic(Id("arr"), indexer.Next(), Qual("C", "int64_t"), Id("o"), Id(t.TypeName())))
		case info&types.IsFloat != 0:
			g.Add(codeLoadBasic(Id("arr"), indexer.Next(), Qual("C", "double"), Id("o"), Id(t.TypeName())))
		case info&types.IsString != 0:
			g.Add(codeLoadStr(Id("arr"), indexer.Next(), Id("o")))
		}
	case *ir.Array:
		Index := indexer.BeginRep(ir.Sizeof(t.Elem), int(t.Len))
		g.For(Index.Clone().Op(":=").Lit(0), Index.Clone().Op("<").Lit(int(t.Len)), Index.Clone().Op("++")).BlockFunc(func(g *Group) {
			g.Add(Id("o").Op(":=").Op("&").Id("o").Index(Index))
			ex.dumpLoadInplaceInit(t.Elem, g, indexer)
		})
	}
}

func (ex *ExportedType) dump() {
	name := ex.Name()
	f := NewFile(ex.pkg.Name)
	f.ImportAlias(dgoMod, "dgo")
	f.CgoPreamble("#include <stdint.h>\n#include <stdbool.h>")
	f.Func().
		Params(Op("*").Id(name)).
		Id("DgoSize").Params().Int().Block(
		Return(Lit(ir.Sizeof(ex.irTerm))),
	)

	f.Func().
		Params(Id("o").Op("*").Id(name)).
		Id("DgoLoadInplace").Params(Id("arr").Index().Op("*").Qual(dgoMod, "Dart_CObject")).BlockFunc(func(g *Group) {
		ex.dumpLoadInplaceInit(ex.irTerm, g, &indexer{})
	})

	filename := ex.pkg.Fset.File(ex.Pos()).Name()
	filename = path.Join(path.Dir(filename), fmt.Sprintf("%s.dgo.go", name))
	err := f.Save(filename)
	Die(err)
}

func (ex *Exports) Dump() {
	for _, epkg := range ex.Pkgs {
		for _, etyp := range epkg.Types {
			etyp.dump()
		}
	}
}
