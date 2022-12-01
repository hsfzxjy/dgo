package ir

func fillSize(t Term, getSize func(Term) *int, skip func(Term) bool) {
	cntSize := 0

	type isdynamic bool
	var stack []isdynamic

	push := func(v isdynamic) { stack = append(stack, v) }
	top := func() *isdynamic {
		if len(stack) == 0 {
			return nil
		} else {
			return &stack[len(stack)-1]
		}
	}
	pop := func() isdynamic {
		x := top()
		stack = stack[:len(stack)-1]
		return *x
	}

	t.Traverse(func(t Term) {
		*getSize(t) = cntSize
		if _, ok := t.(*Basic); ok {
			cntSize++
		}
		push(isdynamic(false))
	}, func(t Term) {
		baseSize := getSize(t)
		if skip(t) {
			cntSize = *baseSize
			*baseSize = 0
			return
		}
		defer func() {
			if pop() == isdynamic(true) {
				*baseSize = -1
				t := top()
				if t != nil {
					*t = isdynamic(true)
				}
			}
		}()
		switch t := t.(type) {
		case *Array:
			arrSize := (cntSize - *baseSize) * int(t.Len)
			cntSize = *baseSize + arrSize
		case *Optional:
			if cntSize == *baseSize {
				cntSize++
			}
		case *Slice, *Map:
			*top() = isdynamic(true)
		}
		*baseSize = cntSize - *baseSize
	})
}

const DYNSIZE = -1

func GoSizeof(t Term) int {
	if t == nil {
		return 0
	}
	return t.getHeader().GoSize
}
func IsGoSizeDynamic(t Term) bool { return GoSizeof(t) == DYNSIZE }
func DartSizeof(t Term) int {
	if t == nil {
		return 0
	}
	return t.getHeader().DartSize
}
func IsDartSizeDynamic(t Term) bool { return DartSizeof(t) == DYNSIZE }

func FillAllSize(t Term) {
	fillSize(t, func(t Term) *int {
		return &t.getHeader().GoSize
	}, func(t Term) bool {
		if t, ok := t.(*Field); ok {
			return !t.SendBackToGo
		}
		return false
	})
	fillSize(t, func(t Term) *int {
		return &t.getHeader().DartSize
	}, func(t Term) bool {
		if t, ok := t.(*Field); ok {
			return !t.SendToDart
		}
		return false
	})
}
