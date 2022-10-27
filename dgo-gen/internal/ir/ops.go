package ir

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
)

type Ident struct {
	Pkg  *types.Package `json:"-"`
	Name string         `json:"-"`
	Uri  uri.Uri
}

type termHeader struct {
	OpName   string `json:"Op"`
	DartSize int
	GoSize   int
}

func (h *termHeader) getHeader() *termHeader { return h }
func (h *termHeader) initHeader(name string) {
	h.OpName = name
	h.DartSize = -1
	h.GoSize = -1
}

type termIdent struct {
	*Ident `json:"Ident,omitempty"`
}

type HasIdent interface {
	GetIdent() *Ident
	SetIdent(*Ident)
}

func (i *termIdent) GetIdent() *Ident      { return i.Ident }
func (i *termIdent) SetIdent(ident *Ident) { i.Ident = ident }

type Array struct {
	termHeader
	termIdent
	Len  int
	Elem Term
}

func NewArray(length int) *Array {
	t := &Array{Len: length}
	t.initHeader("Array")
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
	termIdent
	TypeKind types.BasicKind `json:"-"`
	TypeInfo types.BasicInfo `json:"-"`
	TypeName string
}

func NewBasic(typ *types.Basic) *Basic {
	t := &Basic{TypeKind: typ.Kind(), TypeInfo: typ.Info(), TypeName: typ.Name()}
	t.initHeader("Basic")
	return t
}

func (b *Basic) AddChild(Term)                        { panic("not implemented") }
func (b *Basic) Traverse(visitPre, visitPost visitor) { visitPre.Call(b); visitPost.Call(b) }

func NewIdent(pkg *types.Package, name string, uri uri.Uri) *Ident {
	return &Ident{pkg, name, uri}
}

type Coerce struct {
	termHeader
	*Ident `json:"Target"`
	Elem   Term `json:"-"`
}

func NewCoerce(obj types.Object, uri uri.Uri) *Coerce {
	t := &Coerce{Ident: NewIdent(obj.Pkg(), obj.Name(), uri)}
	t.initHeader("Coerce")
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
	termIdent
	Elem Term
}

func NewPtrTo() *PtrTo {
	t := &PtrTo{}
	t.initHeader("PtrTo")
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

	SendToDart   bool
	SendBackToGo bool
	RenameInDart string
}

func NewField(name string, directive string) (*Field, error) {
	t := &Field{Name: name, SendToDart: true, SendBackToGo: true}

	actions := strings.Split(directive, ",")
	for _, action := range actions {
		switch action {
		case "!dart":
			t.SendToDart = false
			t.SendBackToGo = false
		case "!go":
			t.SendBackToGo = false
		default:
			if t.RenameInDart != "" {
				return nil, fmt.Errorf("multiple renames: %q and %q", t.RenameInDart, action)
			}
			t.RenameInDart = action
		}
	}

	t.initHeader("Field")
	return t, nil
}

func (f *Field) AddChild(t Term) { f.Term = t }
func (f *Field) Traverse(visitPre, visitPost visitor) {
	visitPre.Call(f)
	f.Term.Traverse(visitPre, visitPost)
	visitPost(f)
}

type Struct struct {
	termHeader
	termIdent
	Fields []*Field
}

func NewStruct() *Struct {
	t := &Struct{Fields: []*Field{}}
	t.initHeader("Struct")
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

type Optional struct {
	termHeader
	Term Term
}

func NewOptional() *Optional {
	o := &Optional{}
	o.initHeader("Optional")
	return o
}

func (o *Optional) AddChild(t Term) { o.Term = t }
func (o *Optional) Traverse(visitPre, visitPost visitor) {
	visitPre.Call(o)
	o.Term.Traverse(visitPre, visitPost)
	visitPost.Call(o)
}
