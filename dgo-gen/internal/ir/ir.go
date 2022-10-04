package ir

type visitKind int

const (
	vkInvalid visitKind = iota
	vkEnter
	vkLeave
	vkLeaf
)

type visitor func(Term, visitKind)

type Term interface {
	Traverse(visitor)
	AddChild(Term)
}

type Elem = Term
