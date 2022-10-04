package ir

import (
	"go/types"
)

type Array struct {
	cachedSize
	Len  int64
	Elem Term
}

func (arr *Array) AddChild(t Term) { arr.Elem = t }

func (arr *Array) Traverse(visit visitor) {
	visit(arr, vkEnter)
	arr.Elem.Traverse(visit)
	visit(arr, vkLeave)
}

type Basic struct {
	Type *types.Basic
}

func (b *Basic) AddChild(Term)          { panic("not implemented") }
func (b *Basic) Traverse(visit visitor) { visit(b, vkLeaf) }
func (b *Basic) TypeName() string       { return b.Type.String() }

type Ident struct {
	Pkg  *types.Package
	Name string
}

type Coerce struct {
	cachedSize
	Ident
	Elem
}

func (c *Coerce) AddChild(t Term) { c.Elem = t }
func (c *Coerce) Traverse(visit visitor) {
	visit(c, vkEnter)
	c.Elem.Traverse(visit)
	visit(c, vkLeave)
}

type PtrTo struct {
	cachedSize
	Elem
}

func (p *PtrTo) AddChild(t Term) { p.Elem = t }
func (p *PtrTo) Traverse(visit visitor) {
	visit(p, vkEnter)
	p.Elem.Traverse(visit)
	visit(p, vkLeave)
}

type Field struct {
	cachedSize
	Name string
	Term
}

func (f *Field) AddChild(t Term) { f.Term = t }
func (f *Field) Traverse(visit visitor) {
	visit(f, vkEnter)
	f.Term.Traverse(visit)
	visit(f, vkLeave)
}

type Struct struct {
	cachedSize
	Fields []*Field
}

func (s *Struct) AddChild(t Term) { s.Fields = append(s.Fields, t.(*Field)) }
func (s *Struct) Traverse(visit visitor) {
	visit(s, vkEnter)
	for _, field := range s.Fields {
		field.Traverse(visit)
	}
	visit(s, vkLeave)
}
