package gogen

import (
	"fmt"
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
			if etype.PPackage.PkgPath == t.Pkg.Path() {
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
	return Op("*").Add(dst).Op("=").
		Add(dstType).
		Parens(
			Op("*").
				Parens(Op("*").Add(srcType)).
				Parens(
					Qual("unsafe", "Pointer").Parens(Op("&").Add(src).Index(idx).Dot("Value"))))
}

func loadIntoString(src, idx, dst Code) *Statement {
	return Block(
		Id("pStr").Op(":=").
			Op("*").
			Parens(Op("*").Op("*").Index(Qual(dgoMod, "MAX_ARRAY_LEN")).Byte()).
			Parens(
				Qual("unsafe", "Pointer").
					Parens(Op("&").Add(src).Index(idx).Dot("Value"))),

		Id("length").Op(":=").
			Qual("bytes", "IndexByte").
			Call(
				Id("pStr").Index(Empty(), Empty()),
				LitByte('\x00')),

		Empty().
			If(
				Id("length").Op("<").Lit(0).
					Op("||").
					Id("pStr").Index(Id("length")).Op("!=").LitByte('\x00')).
			Block(
				Panic(Lit("dgo:go string too long"))),

		Id("byteSlice").Op(":=").Make(Index().Byte(), Id("length")),

		Copy(
			Id("byteSlice").Index(Empty(), Id("length")),
			Id("pStr").Index(Empty(), Id("length"))),

		Op("*").Add(dst).Op("=").String().Parens(Id("byteSlice")))
}

func loadIntoInt(src, idx, dst, dstType Code) *Statement {
	return If(Id("arr").Index(Id("_index_")).Dot("Type").
		Op("==").
		Qual(dgoMod, "Dart_CObject_kInt32")).
		BlockFunc(func(g *Group) {
			g.Add(loadIntoBasic(src, idx, Qual("C", "int32_t"), dst, dstType))
		}).
		Else().
		BlockFunc(func(g *Group) {
			g.Add(loadIntoBasic(src, idx, Qual("C", "int64_t"), dst, dstType))
		})
}

func loadBasic(t *ir.Basic, g *Group) {
	info := t.TypeInfo
	switch {
	case info&types.IsBoolean != 0:
		g.Add(loadIntoBasic(Id("arr"), Id("_index_"), Qual("C", "bool"), Id("o"), Id("bool")))
	case info&types.IsInteger != 0:
		g.Add(loadIntoInt(Id("arr"), Id("_index_"), Id("o"), Id(t.TypeName)))
	case info&types.IsFloat != 0:
		g.Add(loadIntoBasic(Id("arr"), Id("_index_"), Qual("C", "double"), Id("o"), Id(t.TypeName)))
	case info&types.IsString != 0:
		g.Add(loadIntoString(Id("arr"), Id("_index_"), Id("o")))
	}
	g.Add(Id("_index_").Op("++"))
}

func buildFunction_DgoLoad(etype *exported.Type, term ir.Term, g *Group, looper *looper) {
SWITCH:
	switch t := term.(type) {
	case *ir.Struct:
		for _, f := range t.Fields {
			if !f.SendBackToGo {
				continue
			}
			buildFunction_DgoLoad(etype, f, g, looper)
		}
	case *ir.Field:
		g.BlockFunc(func(g *Group) {
			g.Id("o").
				Op(":=").
				Op("&").Id("o").Dot(t.Name)
			buildFunction_DgoLoad(etype, t.Term, g, looper)
		})
	case *ir.PtrTo:
		g.Op("*").Id("o").
			Op("=").
			New(typeNameOf(etype, t))
		term = t.Elem
		goto SWITCH
	case *ir.Coerce:
		g.Id("_index_").
			Op("=").
			Id("o").Dot("DgoLoad").
			Call(
				Id("arr"),
				Id("_index_"))
	case *ir.Basic:
		if t.Ident != nil {
			g.BlockFunc(func(g *Group) {
				g.Id("o").
					Op(":=").
					Parens(Op("*").Id(t.TypeName)).
					Parens(Id("o"))
				loadBasic(t, g)
			})
		} else {
			loadBasic(t, g)
		}
	case *ir.Array:
		if t.Len > 0 {
			Looper := looper.BeginRep()
			g.
				For(
					Looper.Clone().Op(":=").Lit(0),
					Looper.Clone().Op("<").Lit(int(t.Len)),
					Looper.Clone().Op("++")).
				BlockFunc(func(g *Group) {
					g.Id("o").
						Op(":=").
						Op("&").Id("o").Index(Looper)
					buildFunction_DgoLoad(etype, t.Elem, g, looper)
				})
			looper.EndRep()
		} else {
			g.Id("_").Op("=").Op("o")
		}
	case *ir.Optional:
		g.
			If(Empty().
				Id("arr").Index(Id("_index_")).Dot("Type").
				Op("==").
				Qual(dgoMod, "Dart_CObject_kNull")).
			Block(
				Id("o").Dot("Clear").Call(),
				Id("_index_").Op("++")).
			Else().
			BlockFunc(func(g *Group) {
				g.Id("o").Dot("SetZero").Call()
				g.Id("o").Op(":=").Op("&").Id("o").Dot("Value")
				buildFunction_DgoLoad(etype, t.Term, g, looper)
			})
	}
}

func storeFromBasic(g *Group, kind, ctype Code) {
	g.Id("cobj").
		Op("=").
		Op("&").Id("arr").Index(Id("_index_"))
	g.Id("cobj").Dot("Type").
		Op("=").
		Add(kind)
	g.
		Op("*").Parens(Op("*").Add(ctype)).
		Parens(Qual("unsafe", "Pointer").Parens(Op("&").Id("cobj").Dot("Value"))).
		Op("=").
		Add(ctype).Parens(Id("*").Id("o"))
}

func storeFromString(g *Group) {
	g.Block(
		Id("o").Op(":=").Op("*").Id("o"),

		Empty().
			If(Len(Id("o")).Op("==").Lit(0).
				Op("||").
				Id("o").Index(Len(Id("o")).Op("-").Lit(1)).Op("!=").LitByte('\x00')).
			Block(
				Id("o").Op("=").Id("o").Op("+").Lit("\x00"),
				Op("*").Id("keepAlive").Op("=").Append(Op("*").Id("keepAlive"), Id("o"))),

		Id("header").
			Op(":=").
			Parens(Op("*").Qual("reflect", "StringHeader")).
			Parens(Qual("unsafe", "Pointer").Parens(Op("&").Id("o"))),

		Id("cobj").
			Op("=").
			Op("&").Id("arr").Index(Id("_index_")),

		Id("cobj").Dot("Type").
			Op("=").
			Add(Qual(dgoMod, "Dart_CObject_kString")),

		Op("*").Parens(Op("*").Id("uintptr")).
			Parens(Qual("unsafe", "Pointer").Parens(Op("&").Id("cobj").Dot("Value"))).
			Op("=").
			Id("header").Dot("Data"))
}

func storeFromInt(g *Group) {
	storeFromBasic(g, Qual(dgoMod, "Dart_CObject_kInt64"), Qual("C", "int64_t"))
}

func storeBasic(t *ir.Basic, g *Group) {
	info := t.TypeInfo
	switch {
	case info&types.IsBoolean != 0:
		storeFromBasic(g, Qual(dgoMod, "Dart_CObject_kBool"), Qual("C", "bool"))
	case info&types.IsInteger != 0:
		storeFromInt(g)
	case info&types.IsFloat != 0:
		storeFromBasic(g, Qual(dgoMod, "Dart_CObject_kDouble"), Qual("C", "double"))
	case info&types.IsString != 0:
		storeFromString(g)
	}
	g.Id("_index_").Op("++")
}

func buildFunction_DgoStore(etype *exported.Type, term ir.Term, g *Group, looper *looper) {
SWITCH:
	switch t := term.(type) {
	case *ir.Struct:
		for _, f := range t.Fields {
			if !f.SendToDart {
				continue
			}
			buildFunction_DgoStore(etype, f, g, looper)
		}
	case *ir.Field:
		g.BlockFunc(func(g *Group) {
			g.Id("o").Op(":=").Op("&").Id("o").Dot(t.Name)
			buildFunction_DgoStore(etype, t.Term, g, looper)
		})
	case *ir.PtrTo:
		g.Op("*").Id("o").Op("=").New(typeNameOf(etype, t))
		term = t.Elem
		goto SWITCH
	case *ir.Coerce:
		g.Id("_index_").
			Op("=").
			Id("o").Dot("DgoStore").
			Call(
				Id("arr").Index(Empty(), Empty()),
				Id("_index_"),
				Id("keepAlive"))
	case *ir.Basic:
		storeBasic(t, g)
	case *ir.Array:
		if t.Len > 0 {
			Looper := looper.BeginRep()
			g.
				For(
					Looper.Clone().Op(":=").Lit(0),
					Looper.Clone().Op("<").Lit(int(t.Len)),
					Looper.Clone().Op("++")).
				BlockFunc(func(g *Group) {
					g.Add(Id("o").Op(":=").Op("&").Id("o").Index(Looper))
					buildFunction_DgoStore(etype, t.Elem, g, looper)
				})
			looper.EndRep()
		} else {
			g.Id("_").Op("=").Op("o")
		}
	case *ir.Optional:
		g.Id("cobj").
			Op("=").
			Op("&").Id("arr").Index(Id("_index_"))
		g.
			If(Id("o").Dot("IsNone").Call()).
			Block(
				Id("cobj").Dot("Type").Op("=").Qual(dgoMod, "Dart_CObject_kNull"),
				Id("_index_").Op("++")).
			Else().
			BlockFunc(func(g *Group) {
				g.Id("o").Op(":=").Op("&").Id("o").Dot("Value")
				buildFunction_DgoStore(etype, t.Term, g, looper)
			})
	}
}

type Generator struct {
	files map[string]*File
}

func (d *Generator) buildFunctionsForType(etype *exported.Type, file *File) {
	name := etype.Name()
	file.Func().
		Params(Op("*").Id(name)).
		Id("DgoGoSize").
		Params().
		Int().
		Block(
			Return(Lit(ir.GoSizeof(etype.Term)))).
		Line()

	file.Func().
		Params(Op("*").Id(name)).
		Id("DgoDartSize").
		Params().
		Int().
		Block(
			Return(Lit(ir.DartSizeof(etype.Term)))).
		Line()

	file.Func().
		Params(Id("o").Op("*").Id(name)).
		Id("DgoLoad").
		Params(
			Id("arr").Index().Op("*").Qual(dgoMod, "Dart_CObject"),
			Id("_index_").Int()).
		Id("int").
		BlockFunc(func(g *Group) {
			buildFunction_DgoLoad(etype, etype.Term, g, new(looper))
			g.Return(Id("_index_"))
		}).
		Line()

	file.Func().
		Params(Id("o").Op("*").Id(name)).
		Id("DgoStore").
		Params(
			Id("arr").Index().Qual(dgoMod, "Dart_CObject"),
			Id("_index_").Id("int"),
			Id("keepAlive").Op("*").Index().Id("any")).
		Id("int").
		BlockFunc(func(g *Group) {
			g.Var().Id("cobj").Op("*").Qual(dgoMod, "Dart_CObject")
			buildFunction_DgoStore(etype, etype.Term, g, new(looper))
			g.Id("_").Op("=").Id("cobj")
			g.Return(Id("_index_"))
		}).Line()

	for _, method := range etype.Methods {
		implName := fmt.Sprintf("dgo_impl_%s_%s", etype.Name(), method.Name)
		file.Func().
			Id(implName).
			Params(
				Id("arr").Index().Op("*").Qual(dgoMod, "Dart_CObject")).
			BlockFunc(func(g *Group) {
				g.Var().Id("cobj").Op("*").Qual(dgoMod, "Dart_CObject")
				g.Id("_index_").Op(":=").Lit(0)
				g.Var().Id("o").Id(etype.Name())
				g.Id("_index_").
					Op("=").
					Id("o").Dot("DgoLoad").
					Call(
						Id("arr"),
						Id("_index_"))

				for paramId, param := range method.Params {
					paramName := fmt.Sprintf("arg%d", paramId)
					g.Var().Id(paramName).Add(typeNameOf(etype, param.Term))
					g.BlockFunc(func(g *Group) {
						g.Id("o").Op(":=").Op("&").Id(paramName)
						buildFunction_DgoLoad(etype, param.Term, g, &looper{})
					})
				}

				g.Var().Id("callback").Uint64()
				g.Add(loadIntoInt(Id("arr"), Id("_index_"), Op("&").Id("callback"), Uint64()))
				g.Id("callback").Op("|=").Uint64().Call(Qual(dgoMod, "CF_POP"))

				var resultReceiver *Statement = Empty()
				if method.Return != nil {
					if method.ReturnError {
						resultReceiver = List(Id("result"), Id("err"))
					} else {
						resultReceiver = Id("result")
					}
					resultReceiver = resultReceiver.Op(":=")
				}

				g.Add(resultReceiver).Id("o").Dot(method.Name).
					CallFunc(func(g *Group) {
						for paramId := range method.Params {
							paramName := fmt.Sprintf("arg%d", paramId)
							g.Id(paramName)
						}
					})

				g.Var().Id("keepAliveArr").Index().Any()
				g.Id("keepAlive").Op(":=").Op("&").Id("keepAliveArr")
				g.Id("_index_").Op("=").Lit(0)
				g.BlockFunc(func(g *Group) {
					if method.Return == nil {
						g.Var().Id("arr").Index(Lit(1)).Qual(dgoMod, "Dart_CObject")
						g.Id("o").Op(":=").Op("&").Id("callback")
						storeFromInt(g)
						g.Id("_index_").Op("++")
					} else if !method.ReturnError {
						g.Var().Id("arr").Index(Lit(1+ir.DartSizeof(method.Return))).Qual(dgoMod, "Dart_CObject")
						g.BlockFunc(func(g *Group) {
							g.Id("o").Op(":=").Op("&").Id("callback")
							storeFromInt(g)
							g.Id("_index_").Op("++")
						})
						g.BlockFunc(func(g *Group) {
							g.Id("o").Op(":=").Op("&").Id("result")
							buildFunction_DgoStore(etype, method.Return, g, &looper{})
						})
					} else {
						g.Var().Id("arr").Index(Lit(2+ir.DartSizeof(method.Return))).Qual(dgoMod, "Dart_CObject")
						g.If(Id("err").Op("!=").Nil()).Block(
							Id("callback").Op("|=").Uint64().Call(Id("cf_fut_reject")),
						)
						g.BlockFunc(func(g *Group) {
							g.Id("o").Op(":=").Op("&").Id("callback")
							storeFromInt(g)
							g.Id("_index_").Op("++")
						})
						g.If(Id("err").Op("!=").Nil()).BlockFunc(func(g *Group) {
							g.Id("errString").Op(":=").Id("err").Dot("Error").Call()
							g.Id("o").Op(":=").Op("&").Id("errString")
							storeFromString(g)
							g.Id("_index_").Op("++")
						}).Else().BlockFunc(func(g *Group) {
							g.Id("o").Op(":=").Op("&").Id("result")
							buildFunction_DgoStore(etype, method.Return, g, &looper{})
						})
					}
					g.Id("dgoPostCObjects").Call(
						Id("_index_"),
						Op("&").Id("arr").Index(Lit(0)),
					)
					g.Qual("runtime", "KeepAlive").Call(Id("keepAlive"))
				})

			}).
			Line().
			Line().
			Func().Id("init").
			Params().
			Block(
				Id("dgoMethodCallRegister").
					Call(
						Qual(dgoMod, "MethodCallId").
							Call(Lit(method.FuncId)),
						Id(implName))).
			Line()
	}
}

func (d *Generator) buildStub(dstDir string, pkgName string) {
	dstPath := path.Join(dstDir, "stub.dgo._.go")
	if _, ok := d.files[dstPath]; ok {
		return
	}
	file := NewFile(pkgName)
	d.files[dstPath] = file
	file.ImportAlias(dgoMod, "dgo")
	file.
		Comment("//go:linkname dgoMethodCallRegister github.com/hsfzxjy/dgo/go.registerMethodCall").
		Line().
		Func().Id("dgoMethodCallRegister").
		Params(
			Qual(dgoMod, "MethodCallId"),
			Qual(dgoMod, "MethodCallImplFunc")).
		Line().
		Line().
		Comment("//go:linkname dgoPostCObjects github.com/hsfzxjy/dgo/go.dgo__PostCObjects").
		Line().
		Func().Id("dgoPostCObjects").
		Params(Int(), Op("*").Qual(dgoMod, "Dart_CObject")).
		Line().
		Line().
		Var().Id("_").Qual("unsafe", "Pointer").
		Line().
		Var().Id("cf_fut_reject").Op("=").Qual(dgoMod, "CF_CUSTOM").Call(Lit(7))
}

func (d *Generator) AddType(etype *exported.Type) {
	if d.files == nil {
		d.files = make(map[string]*File)
	}
	dstDir, dstFilename := path.Split(etype.PPackage.Fset.File(etype.Pos()).Name())
	dstPath := path.Join(dstDir, strings.TrimSuffix(dstFilename, ".go")+".dgo.go")
	var file *File
	{
		var ok bool
		if file, ok = d.files[dstPath]; !ok {
			file = NewFile(etype.PPackage.Name)
			file.ImportAlias(dgoMod, "dgo")
			file.CgoPreamble(cgoPreamble)

			d.files[dstPath] = file

			d.buildStub(dstDir, etype.PPackage.Name)
		}
	}
	d.buildFunctionsForType(etype, file)
}

func (d *Generator) Save() {
	for fileName, file := range d.files {
		err := file.Save(fileName)
		exception.Die(err)
	}
}
