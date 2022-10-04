package internal

import . "github.com/dave/jennifer/jen"

func codeLoadBasic(src, idx, srcType, dst, dstType Code) Code {
	return Op("*").Add(dst).Op("=").Add(dstType).Parens(
		Op("*").Parens(Op("*").Add(srcType)).Parens(
			Qual("unsafe", "Pointer").Parens(Op("&").Add(src).Index(idx).Dot("Value")),
		),
	)
}

func codeLoadStr(src, idx, dst Code) Code {
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
