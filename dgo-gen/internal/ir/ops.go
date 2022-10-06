package ir

import (
	"go/types"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type termHeader struct {
	OpName string `json:"Op"`
	Size   int
}

func (h *termHeader) getHeader() *termHeader { return h }

type Array struct {
	termHeader
	Len  int
	Elem Term
}

func NewArray(length int) *Array {
	t := &Array{Len: length}
	t.OpName = "Array"
	return t
}

func (arr *Array) AddChild(t Term) { arr.Elem = t }

func (arr *Array) Traverse(visitPre, visitPost visitor) {
	visitPre.Call(arr)
	arr.Elem.Traverse(visitPre, visitPost)
	visitPost.Call(arr)
}

type Basic struct {
	termHeader
	TypeKind types.BasicKind `json:"-"`
	TypeInfo types.BasicInfo `json:"-"`
	TypeName string
}

func NewBasic(typ *types.Basic) *Basic {
	t := &Basic{TypeKind: typ.Kind(), TypeInfo: typ.Info(), TypeName: typ.Name()}
	t.OpName = "Basic"
	return t
}

func (b *Basic) AddChild(Term)                        { panic("not implemented") }
func (b *Basic) Traverse(visitPre, visitPost visitor) { visitPre.Call(b); visitPost.Call(b) }

type Ident struct {
	Pkg  *types.Package `json:"-"`
	Name string
	Uri  uri.Uri
}

func NewIdent(pkg *types.Package, name string) *Ident {
	return &Ident{pkg, name, uri.NewUri(pkg.Path(), name)}
}

type Coerce struct {
	termHeader
	*Ident
	Elem Term
}

func NewCoerce(pkg *types.Package, name string) *Coerce {
	t := &Coerce{Ident: NewIdent(pkg, name)}
	t.OpName = "Coerce"
	return t
}

func (c *Coerce) AddChild(t Term) { c.Elem = t }
func (c *Coerce) Traverse(visitPre, visitPost visitor) {
	visitPre.Call(c)
	c.Elem.Traverse(visitPre, visitPost)
	visitPost.Call(c)
}

type PtrTo struct {
	termHeader
	Elem Term
}

func NewPtrTo() *PtrTo {
	t := &PtrTo{}
	t.OpName = "PtrTo"
	return t
}

func (p *PtrTo) AddChild(t Term) { p.Elem = t }
func (p *PtrTo) Traverse(visitPre, visitPost visitor) {
	visitPre.Call(p)
	p.Elem.Traverse(visitPre, visitPost)
	visitPost(p)
}

type Field struct {
	termHeader
	Name string
	Term Term
}

func NewField(name string) *Field {
	t := &Field{Name: name}
	t.OpName = "Field"
	return t
}

func (f *Field) AddChild(t Term) { f.Term = t }
func (f *Field) Traverse(visitPre, visitPost visitor) {
	visitPre.Call(f)
	f.Term.Traverse(visitPre, visitPost)
	visitPost(f)
}

type Struct struct {
	termHeader
	Fields []*Field
}

func NewStruct() *Struct {
	t := &Struct{}
	t.OpName = "Struct"
	return t
}

func (s *Struct) AddChild(t Term) { s.Fields = append(s.Fields, t.(*Field)) }
func (s *Struct) Traverse(visitPre, visitPost visitor) {
	visitPre(s)
	for _, field := range s.Fields {
		field.Traverse(visitPre, visitPost)
	}
	visitPost.Call(s)
}
