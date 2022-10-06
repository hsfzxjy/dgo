package ir

type visitor func(Term)

func (v visitor) Call(term Term) {
	if v != nil {
		v(term)
	}
}

type Term interface {
	Traverse(visitPre, visitPost visitor)
	AddChild(Term)
	getHeader() *termHeader
}

type Elem = Term
