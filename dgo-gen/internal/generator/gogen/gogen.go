package gogen

import (
	"fmt"
	"go/types"
	"os"
	"path"
	"path/filepath"
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

	var ident *ir.Ident
	if t, ok := term.(ir.HasIdent); ok && t.GetIdent() != nil {
		ident = t.GetIdent()
		goto HANDLE_IDENT
	}
LOOP:
	for {
		switch t := term.(type) {
		case *ir.PtrTo:
			s.Op("*")
			term = t.Elem
		case *ir.Array:
			s.Index(Lit(t.Len))
			term = t.Elem
		case *ir.Slice:
			s.Index()
			term = t.Elem
		case *ir.Map:
			s.Map(typeNameOf(etype, t.Key))
			term = t.Value
		case *ir.Coerce:
			ident = t.Ident
			goto HANDLE_IDENT
		case *ir.Basic:
			s.Id(t.TypeName)
			break LOOP
		case *ir.Struct:
			break LOOP
		case *ir.PinToken:
			s.Add(pinToken(typeNameOf(etype, t.Term)))
			break LOOP
		default:
			panic("unreachable")
		}
	}
	goto RETURN

HANDLE_IDENT:
	if etype.PPackage.PkgPath == ident.Pkg.Path() {
		s.Id(ident.Name)
	} else {
		s.Qual(ident.Pkg.Path(), ident.Name)
	}

RETURN:
	return &s
}

func incIndex() *Statement {
	return Id("_index_").Op("++")
}

func arrIndex() *Statement {
	return Id("arr").Index(Id("_index_"))
}

func untypedPinToken() *Statement {
	return pinToken(Struct())
}

func pinToken(typeParam Code) *Statement {
	return Qual(dgoMod+"/pin", "Token").Index(typeParam)
}

func loadIntoBasic(g *Group, src, srcType, dst, dstType Code) {
	g.Add(dst).Op("=").
		Add(dstType).
		Parens(
			Op("*").
				Parens(Op("*").Add(srcType)).
				Parens(
					Qual("unsafe", "Pointer").Parens(Op("&").Add(src).Dot("Value"))))
	g.Add(incIndex())
}

func loadIntoString(g *Group, src, dst Code) {
	g.Block(
		Id("pStr").Op(":=").
			Op("*").
			Parens(Op("*").Op("*").Index(Qual(dgoMod, "MAX_ARRAY_LEN")).Byte()).
			Parens(
				Qual("unsafe", "Pointer").
					Parens(Op("&").Add(src).Dot("Value"))),

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
				Panic(Lit("dgo:go: string too long"))),

		Id("byteSlice").Op(":=").Make(Index().Byte(), Id("length")),

		Copy(
			Id("byteSlice").Index(Empty(), Id("length")),
			Id("pStr").Index(Empty(), Id("length"))),

		Add(dst).Op("=").String().Parens(Id("byteSlice")))
	g.Add(incIndex())
}

func loadIntoInt(g *Group, src, dst, dstType Code) {
	g.Block(
		Id("_obj_").Op(":=").Add(src),
		If(Id("_obj_").Dot("Type").
			Op("==").
			Qual(dgoMod, "Dart_CObject_kInt32")).
			BlockFunc(func(g *Group) {
				loadIntoBasic(g, Id("_obj_"), Qual("C", "int32_t"), dst, dstType)
			}).
			Else().
			BlockFunc(func(g *Group) {
				loadIntoBasic(g, Id("_obj_"), Qual("C", "int64_t"), dst, dstType)
			}),
	)
}

func loadBasic(t *ir.Basic, g *Group, holder Code) {
	info := t.TypeInfo
	switch {
	case info&types.IsBoolean != 0:
		loadIntoBasic(g, arrIndex(), Qual("C", "bool"), holder, Id("bool"))
	case info&types.IsInteger != 0:
		loadIntoInt(g, arrIndex(), holder, Id(t.TypeName))
	case info&types.IsFloat != 0:
		loadIntoBasic(g, arrIndex(), Qual("C", "double"), holder, Id(t.TypeName))
	case info&types.IsString != 0:
		loadIntoString(g, arrIndex(), holder)
	}
}

func buildFunction_DgoLoad(etype *exported.Type, term ir.Term, g *Group, looper *looper) {
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
			New(typeNameOf(etype, t.Elem))
		g.BlockFunc(func(g *Group) {
			g.Id("o").
				Op(":=").Op("*").Id("o")
			buildFunction_DgoLoad(etype, t.Elem, g, looper)
		})
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
				loadBasic(t, g, Op("*o"))
			})
		} else {
			loadBasic(t, g, Op("*o"))
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
	case *ir.Slice:
		g.Var().Id("size").Int()
		loadIntoInt(g, arrIndex(), Id("size"), Int())
		g.Op("*").Id("o").Op("=").Make(typeNameOf(etype, t), Id("size"))
		Looper := looper.BeginRep()
		g.
			For(
				Looper.Clone().Op(":=").Lit(0),
				Looper.Clone().Op("<").Id("size"),
				Looper.Clone().Op("++")).
			BlockFunc(func(g *Group) {
				g.Id("o").
					Op(":=").
					Op("&").Parens(Op("*").Id("o")).Index(Looper)
				buildFunction_DgoLoad(etype, t.Elem, g, looper)
			})
		looper.EndRep()
	case *ir.Map:
		g.Var().Id("size").Int()
		loadIntoInt(g, arrIndex(), Id("size"), Int())
		g.Op("*").Id("o").Op("=").Make(typeNameOf(etype, t), Id("size"))
		Looper := looper.BeginRep()
		g.
			For(
				Looper.Clone().Op(":=").Lit(0),
				Looper.Clone().Op("<").Id("size"),
				Looper.Clone().Op("++")).
			Block(
				Var().Id("key").Add(typeNameOf(etype, t.Key)),
				BlockFunc(func(g *Group) {
					g.Id("o").Op(":=&").Id("key")
					buildFunction_DgoLoad(etype, t.Key, g, looper)
				}),
				Var().Id("value").Add(typeNameOf(etype, t.Value)),
				BlockFunc(func(g *Group) {
					g.Id("o").Op(":=&").Id("value")
					buildFunction_DgoLoad(etype, t.Value, g, looper)
				}),
				Parens(Op("*").Id("o")).Index(Id("key")).Op("=").Id("value"),
			)
		looper.EndRep()
	case *ir.Optional:
		g.
			If(Empty().
				Id("arr").Index(Id("_index_")).Dot("Type").
				Op("==").
				Qual(dgoMod, "Dart_CObject_kNull")).
			Block(
				Id("o").Dot("Clear").Call(),
				incIndex()).
			Else().
			BlockFunc(func(g *Group) {
				g.Id("o").Dot("SetZero").Call()
				g.Id("o").Op(":=").Op("&").Id("o").Dot("Value")
				buildFunction_DgoLoad(etype, t.Term, g, looper)
			})
	case *ir.PinToken:
		g.BlockFunc(func(g *Group) {
			g.Var().Id("version").Uint16()
			g.Var().Id("data").Uintptr()
			loadIntoInt(g, arrIndex(), Id("version"), Uint16())
			loadIntoInt(g, arrIndex(), Id("data"), Uintptr())
			g.Op("*o").Op("=").
				Add(Parens(typeNameOf(etype, t))).
				Parens(
					Id("pin_untypedTokenFromRaw").
						Call(Id("version"), Id("data")),
				)
		})
	}
}

func storeFromBasic(g *Group, holder, kind, ctype Code) {
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
		Add(ctype).Parens(holder)
	g.Add(incIndex())
}

func storeFromString(g *Group, holder Code) {
	g.Block(
		Id("o").Op(":=").Add(holder),

		Empty().
			If(Len(Id("o")).Op("==").Lit(0).
				Op("||").
				Id("o").Index(Len(Id("o")).Op("-").Lit(1)).Op("!=").LitByte('\x00')).
			Block(
				Id("o").Op("=").Id("o").Op("+").Lit("\x00")),

		Op("*").Id("keepAlive").Op("=").Append(Op("*").Id("keepAlive"), Id("o")),

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
	g.Add(incIndex())
}

func storeFromInt(g *Group, holder Code) {
	storeFromBasic(g, holder, Qual(dgoMod, "Dart_CObject_kInt64"), Qual("C", "int64_t"))
}

func storeBasic(t *ir.Basic, g *Group, holder Code) {
	info := t.TypeInfo
	switch {
	case info&types.IsBoolean != 0:
		storeFromBasic(g, holder, Qual(dgoMod, "Dart_CObject_kBool"), Qual("C", "bool"))
	case info&types.IsInteger != 0:
		storeFromInt(g, holder)
	case info&types.IsFloat != 0:
		storeFromBasic(g, holder, Qual(dgoMod, "Dart_CObject_kDouble"), Qual("C", "double"))
	case info&types.IsString != 0:
		storeFromString(g, holder)
	}
}

func buildFunction_DgoStore(etype *exported.Type, term ir.Term, g *Group, looper *looper) {
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
		g.Op("*").Id("o").Op("=").New(typeNameOf(etype, t.Elem))
		g.BlockFunc(func(g *Group) {
			g.Id("o").Op(":=").Op("*").Id("o")
			buildFunction_DgoStore(etype, t.Elem, g, looper)
		})
	case *ir.Coerce:
		g.Id("_index_").
			Op("=").
			Id("o").Dot("DgoStore").
			Call(
				Id("arr").Index(Empty(), Empty()),
				Id("_index_"),
				Id("keepAlive"))
	case *ir.Basic:
		storeBasic(t, g, Op("*o"))
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
	case *ir.Slice:
		g.BlockFunc(func(g *Group) {
			g.Id("size").Op(":=").Len(Op("*").Id("o"))
			storeFromInt(g, Id("size"))
		})
		Looper := looper.BeginRep()
		g.
			For(
				Looper.Clone().Op(":=").Lit(0),
				Looper.Clone().Op("<").Len(Op("*").Id("o")),
				Looper.Clone().Op("++")).
			BlockFunc(func(g *Group) {
				g.Add(Id("o").Op(":=").Op("&").Parens(Op("*").Id("o")).Index(Looper))
				buildFunction_DgoStore(etype, t.Elem, g, looper)
			})
		looper.EndRep()
	case *ir.Map:
		g.BlockFunc(func(g *Group) {
			g.Id("size").Op(":=").Len(Op("*").Id("o"))
			storeFromInt(g, Id("size"))
		})
		Looper := looper.BeginRep()
		g.
			For(List(Id("key"), Looper).Op(":=").Range().Op("*").Id("o")).
			Block(
				BlockFunc(func(g *Group) {
					g.Add(Id("o").Op(":=").Op("&").Id("key"))
					buildFunction_DgoStore(etype, t.Key, g, looper)
				}),
				BlockFunc(func(g *Group) {
					g.Add(Id("o").Op(":=").Op("&").Add(Looper))
					buildFunction_DgoStore(etype, t.Value, g, looper)
				}),
			)
		looper.EndRep()
	case *ir.Optional:
		g.Id("cobj").
			Op("=").
			Op("&").Id("arr").Index(Id("_index_"))
		g.
			If(Id("o").Dot("IsNone").Call()).
			Block(
				Id("cobj").Dot("Type").Op("=").Qual(dgoMod, "Dart_CObject_kNull"),
				incIndex()).
			Else().
			BlockFunc(func(g *Group) {
				g.Id("o").Op(":=").Op("&").Id("o").Dot("Value")
				buildFunction_DgoStore(etype, t.Term, g, looper)
			})
	case *ir.PinToken:
		g.BlockFunc(func(g *Group) {
			g.List(Id("version"), Id("data")).Op(":=").
				Id("pin_untypedTokenExtract").
				Call(untypedPinToken().Call(Op("*o")))
			storeFromInt(g, Id("version"))
			storeFromInt(g, Id("data"))
			g.Id("o").Op(":=").Id("o").Dot("Data").Call()
			buildFunction_DgoStore(etype, t.Term, g, looper)
		})
		g.Id("pin_untypedTokenLeak").Call(Id("pin_untypedToken").Call(Op("*o")))
	}
}

func buildFunction_method(etype *exported.Type, method exported.TypeMethod, g *Group) {
	g.Var().Id("cobj").Op("*").Qual(dgoMod, "Dart_CObject")
	g.Id("_index_").Op(":=").Lit(0)

	g.Var().Id("callback").Uint64()
	loadIntoBasic(g, arrIndex(), Qual("C", "uint64_t"), Id("callback"), Uint64())
	g.Id("callback").Op("|=").Uint64().Call(Qual(dgoMod, "CF_POP"))
	if method.Return != nil {
		g.Id("callback").Op("|=").Uint64().Call(Qual(dgoMod, "CF_PACKARRAY"))
	}

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

	var resultReceiver *Statement = &Statement{}
	switch {
	case method.Return != nil && method.ReturnError:
		resultReceiver.List(Id("result"), Id("err"))
	case method.Return != nil:
		resultReceiver.Id("result")
	case method.ReturnError:
		resultReceiver.Id("err")
	}
	if len(*resultReceiver) != 0 {
		resultReceiver.Op(":=")
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

	defineArrAndStoreCallback := func(g *Group, nExtraArgs int) {
		r := method.Return
		if ir.IsDartSizeDynamic(r) {
			g.Id("size").Op(":=").Lit(nExtraArgs)
			g.BlockFunc(func(g *Group) {
				g.Id("o").Op(":=&").Id("result")
				buildFunction_DgoDartSize(etype, r, g, new(looper))
			})
			g.Id("arr").Op(":=").Make(Index().Qual(dgoMod, "Dart_CObject"), Id("size"))
		} else {
			g.Var().Id("arr").Index(Lit(nExtraArgs+ir.DartSizeof(r))).Qual(dgoMod, "Dart_CObject")
		}
		g.BlockFunc(func(g *Group) {
			storeFromInt(g, Id("callback"))
		})
	}

	storeErr := func(g *Group) {
		g.If(Id("err").Op("!=").Nil()).
			BlockFunc(func(g *Group) {
				g.Id("errString").Op(":=").Id("err").Dot("Error").Call()
				storeFromString(g, Id("errString"))
			}).
			Else().BlockFunc(func(g *Group) {
			g.Id("arr").Index(Id("_index_")).Dot("Type").
				Op("=").Qual(dgoMod, "Dart_CObject_kNull")
			g.Add(incIndex())
		})
	}

	storeResult := func(g *Group) {
		g.Id("o").Op(":=").Op("&").Id("result")
		buildFunction_DgoStore(etype, method.Return, g, &looper{})
	}

	g.BlockFunc(func(g *Group) {
		switch {
		case method.Return != nil && method.ReturnError:
			defineArrAndStoreCallback(g, 2)
			storeErr(g)
			storeResult(g)
		case method.Return != nil:
			defineArrAndStoreCallback(g, 1)
			storeResult(g)
		case method.ReturnError:
			defineArrAndStoreCallback(g, 2)
			storeErr(g)
		default:
			defineArrAndStoreCallback(g, 1)
		}

		g.Id("dgoPostCObjects").Call(
			Id("port"),
			Id("_index_"),
			Op("&").Id("arr").Index(Lit(0)),
		)
		g.Qual("runtime", "KeepAlive").Call(Id("keepAlive"))
	})
}

func buildFunction_DgoDartSize(etype *exported.Type, term ir.Term, g *Group, looper *looper) {
	if !ir.IsDartSizeDynamic(term) {
		g.Id("size").Op("+=").Lit(ir.GoSizeof(term))
		g.Id("_").Op("=").Id("o")
		return
	}
	switch t := term.(type) {
	case *ir.Struct:
		for _, f := range t.Fields {
			if !f.SendToDart {
				continue
			}
			buildFunction_DgoDartSize(etype, f, g, looper)
		}
	case *ir.Field:
		g.BlockFunc(func(g *Group) {
			g.Id("o").Op(":=").Op("&").Id("o").Dot(t.Name)
			buildFunction_DgoDartSize(etype, t.Term, g, looper)
		})
	case *ir.PtrTo:
		g.BlockFunc(func(g *Group) {
			g.Id("o").Op(":=").Op("*").Id("o")
			buildFunction_DgoDartSize(etype, t.Elem, g, looper)
		})
	case *ir.Coerce:
		g.Id("size").Op("+=").Id("o").Dot("DgoDartSize").Call()
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
					buildFunction_DgoDartSize(etype, t.Elem, g, looper)
				})
			looper.EndRep()
		} else {
			g.Id("_").Op("=").Op("o")
		}
	case *ir.Slice:
		if !ir.IsDartSizeDynamic(t.Elem) {
			g.Id("size").Op("+=").Len(Op("*").Id("o")).Op("*").Lit(ir.DartSizeof(t.Elem)).Op("+").Lit(1)
			return
		}
		Looper := looper.BeginRep()
		g.
			For(
				Looper.Clone().Op(":=").Lit(0),
				Looper.Clone().Op("<").Len(Op("*").Id("o")),
				Looper.Clone().Op("++")).
			BlockFunc(func(g *Group) {
				g.Add(Id("o").Op(":=").Op("&").Parens(Op("*").Id("o")).Index(Looper))
				buildFunction_DgoDartSize(etype, t.Elem, g, looper)
			})
		looper.EndRep()
		g.Id("size").Op("++")
	case *ir.Map:
		if ir.IsDartSizeDynamic(t.Key) {
			panic("dynamic map key")
		}
		if !ir.IsDartSizeDynamic(t.Value) {
			g.Id("size").Op("+=").Len(Op("*").Id("o")).Op("*").
				Lit(ir.DartSizeof(t.Key) + ir.DartSizeof(t.Value)).
				Op("+").Lit(1)
			return
		}
		Looper := looper.BeginRep()
		g.
			For(List(Id("_"), Looper).Op(":=").Range().Op("*").Id("o")).
			BlockFunc(func(g *Group) {
				g.Id("size").Op("+=").Lit(ir.DartSizeof(t.Key))
				g.Id("o").Op(":=&").Add(Looper)
				buildFunction_DgoDartSize(etype, t.Value, g, looper)
			})
		looper.EndRep()
		g.Id("size").Op("++")
	case *ir.Optional:
		g.If(Id("o").Dot("IsSome").Call()).
			BlockFunc(func(g *Group) {
				g.Id("o").Op(":=").Op("&").Id("o").Dot("Value")
				buildFunction_DgoDartSize(etype, t.Term, g, looper)
			}).Else().
			Block(Id("size").Op("+=").Lit(1))
	}
}

type Generator struct {
	files map[string]*File
}

func (d *Generator) buildFunctionsForType(etype *exported.Type, file *File) {
	name := etype.Name()
	typeName := typeNameOf(etype, etype.Term)

	file.Commentf("/*** GENERATED CODE FOR %s ***/", name).
		Line()

	if etype.IsPinnable {
		file.Func().
			Params(Id("o").Op("*").Id(name)).
			Id("NewToken").
			Params().
			Add(pinToken(typeName)).
			Block(Empty().
				If(Qual("unsafe", "Pointer").Call(Id("o")).
					Op("!=").Qual("unsafe", "Pointer").
					Call(Op("&").Id("o").Dot("Meta"))).
				Block(
					Panic(Lit("dgo:go: pinnable object must have an embedded"+
						" pin.Meta as the first field"))),

				Return(Parens(pinToken(typeName)).
					Parens(Id("pin_metaNewToken").Call(
						Op("&").Id("o").Dot("Meta"))),
				))
	}

	file.Func().
		Params(Id("o").Op("*").Id(name)).
		Id("DgoDartSize").
		Params().
		Parens(List(Id("size").Int())).
		BlockFunc(func(g *Group) {
			buildFunction_DgoDartSize(etype, etype.Term, g, new(looper))
			g.Return()
		}).
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
				Id("port").Op("*").Qual(dgoMod, "Port"),
				Id("arr").Index().Op("*").Qual(dgoMod, "Dart_CObject")).
			BlockFunc(func(g *Group) {
				buildFunction_method(etype, method, g)
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
		Comment("//go:linkname dgoMethodCallRegister github.com/hsfzxjy/dgo/go.methodCallRegister").
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
		Params(Op("*").Qual(dgoMod, "Port"), Int(), Op("*").Qual(dgoMod, "Dart_CObject")).
		Line().
		Line().
		Var().Id("_").Qual("unsafe", "Pointer")

	file.
		Type().Id("pin_untypedToken").Op("=").Add(untypedPinToken()).
		Line().
		Line().
		Comment("//go:linkname pin_metaNewToken github.com/hsfzxjy/dgo/go/pin.metaNewToken").
		Line().
		Func().Id("pin_metaNewToken").
		Params(Id("meta").Op("*").Qual(dgoMod+"/pin", "Meta")).
		Add(untypedPinToken()).
		Line().
		Line().
		Comment("//go:linkname pin_untypedTokenFromRaw github.com/hsfzxjy/dgo/go/pin.untypedTokenFromRaw").
		Line().
		Func().Id("pin_untypedTokenFromRaw").
		Params(Id("version").Uint16(), Id("data").Uintptr()).
		Id("pin_untypedToken").
		Line().
		Line().
		Comment("//go:linkname pin_untypedTokenLeak github.com/hsfzxjy/dgo/go/pin.untypedTokenLeak").
		Line().
		Func().Id("pin_untypedTokenLeak").
		Params(Id("token").Id("pin_untypedToken")).
		Line().
		Line().
		Comment("//go:linkname pin_untypedTokenExtract github.com/hsfzxjy/dgo/go/pin.untypedTokenExtract").
		Line().
		Func().Id("pin_untypedTokenExtract").
		Params(Id("token").Add(untypedPinToken())).
		Parens(List(Id("version").Uint16(), Id("data").Uintptr())).
		Line()
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

func (d *Generator) clearOldGeneration() {
	visited := make(map[string]struct{})
	for filename := range d.files {
		dir := path.Dir(filename)
		if _, ok := visited[dir]; ok {
			continue
		}
		paths, err := filepath.Glob(path.Join(dir, "*.dgo.go"))
		if err != nil {
			continue
		}
		for _, path := range paths {
			os.Remove(path)
		}
	}
}

func (d *Generator) Save() {
	d.clearOldGeneration()
	for fileName, file := range d.files {
		err := file.Save(fileName)
		exception.Die(err)
	}
}
