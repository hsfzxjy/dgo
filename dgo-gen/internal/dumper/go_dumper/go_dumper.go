package go_dumper

import (
	"go/types"
	"path"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exception"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/exported"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
)

const dgoMod = "github.com/hsfzxjy/dgo/go"
const cgoPreamble = "#include <stdint.h>\n#include <stdbool.h>"

func typeNameOf(etype *exported.Type, term ir.Term) *Statement {
	s := Statement(nil)
LOOP:
	for {
		switch t := term.(type) {
		case *ir.PtrTo:
			s.Op("*")
			term = t.Elem
		case *ir.Array:
			s.Index(Lit(t.Len))
			term = t.Elem
		case *ir.Coerce:
			if etype.OriPkg.PkgPath == t.Pkg.Path() {
				s.Id(t.Name)
			} else {
				s.Qual(t.Pkg.Path(), t.Name)
			}
			term = t.Elem
		case *ir.Basic:
			s.Id(t.TypeName)
			break LOOP
		case *ir.Struct:
			break LOOP
		default:
			panic("unreachable")
		}
	}
	return &s
}

func loadIntoBasic(src, idx, srcType, dst, dstType Code) *Statement {
	return Op("*").Add(dst).Op("=").Add(dstType).Parens(
		Op("*").Parens(Op("*").Add(srcType)).Parens(
			Qual("unsafe", "Pointer").Parens(Op("&").Add(src).Index(idx).Dot("Value")),
		),
	)
}

func loadIntoString(src, idx, dst Code) *Statement {
	return Block(
		Id("pStr").Op(":=").Op("*").Parens(
			Op("*").Op("*").Index(Qual(dgoMod, "MAX_ARRAY_LEN")).Byte(),
		).Parens(
			Qual("unsafe", "Pointer").Parens(
				Op("&").Add(src).Index(idx).Dot("Value"),
			),
		),
		Id("length").Op(":=").Qual("bytes", "IndexByte").Call(
			Id("pStr").Index(Empty(), Empty()),
			LitByte('\x00'),
		),
		If(Id("length").Op("<").Lit(0).Op("||").Id("pStr").Index(Id("length")).Op("!=").LitByte('\x00')).Block(
			Panic(Lit("dgo:go string too long")),
		),
		Id("byteSlice").Op(":=").Make(Index().Byte(), Id("length")),
		Copy(Id("byteSlice").Index(Empty(), Id("length")), Id("pStr").Index(Empty(), Id("length"))),
		Op("*").Add(dst).Op("=").String().Parens(Id("byteSlice")),
	)
}

func loadInplace(etype *exported.Type, term ir.Term, g *Group, indexer *indexer) {
SWITCH:
	switch t := term.(type) {
	case *ir.Struct:
		for _, f := range t.Fields {
			loadInplace(etype, f, g, indexer)
		}
	case *ir.Field:
		g.BlockFunc(func(g *Group) {
			g.Id("o").Op(":=").Op("&").Id("o").Dot(t.Name)
			loadInplace(etype, t.Term, g, indexer)
		})
	case *ir.PtrTo:
		g.Op("*").Id("o").Op("=").New(typeNameOf(etype, t))
		term = t.Elem
		goto SWITCH
	case *ir.Coerce:
		if el, ok := t.Elem.(*ir.Basic); ok {
			g.BlockFunc(func(g *Group) {
				g.Id("o").Op(":=").Parens(Op("*").Id(el.TypeName)).Parens(Id("o"))
				loadInplace(etype, t.Elem, g, indexer)
			})
		} else {
			term = t.Elem
			goto SWITCH
		}
	case *ir.Basic:
		info := t.TypeInfo
		switch {
		case info&types.IsBoolean != 0:
			g.Add(loadIntoBasic(Id("arr"), indexer.Next(), Qual("C", "bool"), Id("o"), Id("bool")))
		case info&types.IsInteger != 0:
			g.Add(loadIntoBasic(Id("arr"), indexer.Next(), Qual("C", "int64_t"), Id("o"), Id(t.TypeName)))
		case info&types.IsFloat != 0:
			g.Add(loadIntoBasic(Id("arr"), indexer.Next(), Qual("C", "double"), Id("o"), Id(t.TypeName)))
		case info&types.IsString != 0:
			g.Add(loadIntoString(Id("arr"), indexer.Next(), Id("o")))
		}
	case *ir.Array:
		Index := indexer.BeginRep(ir.Sizeof(t.Elem), int(t.Len))
		g.For(Index.Clone().Op(":=").Lit(0), Index.Clone().Op("<").Lit(int(t.Len)), Index.Clone().Op("++")).BlockFunc(func(g *Group) {
			g.Add(Id("o").Op(":=").Op("&").Id("o").Index(Index))
			loadInplace(etype, t.Elem, g, indexer)
		})
	}
}

type GoDumper struct {
	files map[string]*File
}

func (d *GoDumper) dumpType(etype *exported.Type, file *File) {
	name := etype.Name()
	file.Func().
		Params(Op("*").Id(name)).
		Id("DgoSize").Params().Int().
		Block(Return(Lit(ir.Sizeof(etype.IrTerm)))).
		Line()

	file.Func().
		Params(Id("o").Op("*").Id(name)).
		Id("DgoLoadInplace").Params(Id("arr").Index().Op("*").Qual(dgoMod, "Dart_CObject")).
		BlockFunc(func(g *Group) {
			loadInplace(etype, etype.IrTerm, g, new(indexer))
		}).
		Line()
}

func (d *GoDumper) AddType(etype *exported.Type) {
	if d.files == nil {
		d.files = make(map[string]*File)
	}
	dstDir, dstFilename := path.Split(etype.OriPkg.Fset.File(etype.Pos()).Name())
	dstPath := path.Join(dstDir, strings.TrimSuffix(dstFilename, ".go")+".dgo.go")
	var file *File
	{
		var ok bool
		if file, ok = d.files[dstPath]; !ok {
			file = NewFile(etype.OriPkg.Name)
			file.ImportAlias(dgoMod, "dgo")
			file.CgoPreamble(cgoPreamble)

			d.files[dstPath] = file
		}
	}
	d.dumpType(etype, file)
}

func (d *GoDumper) Save() {
	for fileName, file := range d.files {
		err := file.Save(fileName)
		exception.Die(err)
	}
}
