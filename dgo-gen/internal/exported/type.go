package exported

import (
	"fmt"
	"go/types"

	"github.com/hsfzxjy/dgo/dgo-gen/internal/collector"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/ir"
	"github.com/hsfzxjy/dgo/dgo-gen/internal/uri"
	"golang.org/x/tools/go/packages"
)

type TypeMethod struct {
	Name string
	*Function
}

type EnumMember struct{ Name, Value string }

type Type struct {
	types.Object       `json:"-"`
	*collector.Context `json:"-"`
	PPackage           *packages.Package `json:"-"`
	Term               ir.Term
	TypeId             int64
	Methods            []TypeMethod
	IsEnum             bool
	EnumMembers        []EnumMember
	Rename             string

	name2EnumMember  map[string]*EnumMember `json:"-"`
	value2EnumMember map[string]*EnumMember `json:"-"`
}

func (t *Type) parseDirectives(directives []string) {
	if len(directives) == 0 {
		return
	}
	t.Rename = directives[0]
	for _, directive := range directives[1:] {
		if directive == "enum" {
			t.IsEnum = true
		}
	}
}

func (t *Type) addEnumMember(name, value string) error {
	if name == "-" {
		return nil
	}
	if t.name2EnumMember == nil {
		t.name2EnumMember = make(map[string]*EnumMember)
		t.value2EnumMember = make(map[string]*EnumMember)
	}
	if m, ok := t.name2EnumMember[name]; ok {
		return fmt.Errorf("enum member name conflict: %s=%s, %s=%s", m.Name, m.Value, name, value)
	}
	if m, ok := t.value2EnumMember[value]; ok {
		return fmt.Errorf("enum member value conflict: %s=%s, %s=%s", m.Name, m.Value, name, value)
	}
	t.EnumMembers = append(t.EnumMembers, EnumMember{name, value})
	current := &t.EnumMembers[len(t.EnumMembers)-1]
	t.name2EnumMember[name] = current
	t.value2EnumMember[value] = current
	return nil
}

func (t *Type) Uri() uri.Uri {
	return uri.UriFor(t.PPackage, t.Name())
}

func (t *Type) Resolve() {
	t.Term = t.SolveType(t)
}
