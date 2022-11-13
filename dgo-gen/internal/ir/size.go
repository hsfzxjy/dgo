package ir

func fillSize(t Term, getSize func(Term) *int, skip func(Term) bool) {
	cntSize := 0
	t.Traverse(func(t Term) {
		*getSize(t) = cntSize
		if _, ok := t.(*Basic); ok {
			cntSize++
		}
	}, func(t Term) {
		baseSize := getSize(t)
		if skip(t) {
			cntSize = *baseSize
			*baseSize = 0
			return
		}
		switch t := t.(type) {
		case *Array:
			arrSize := (cntSize - *baseSize) * int(t.Len)
			cntSize = *baseSize + arrSize
		case *Optional:
			if cntSize == *baseSize {
				cntSize++
			}
		}
		*baseSize = cntSize - *baseSize
	})
}

func GoSizeof(t Term) int {
	header := t.getHeader()
	if header.GoSize >= 0 {
		return header.GoSize
	}

	fillSize(t, func(t Term) *int {
		return &t.getHeader().GoSize
	}, func(t Term) bool {
		if t, ok := t.(*Field); ok {
			return !t.SendBackToGo
		}
		return false
	})
	return header.GoSize
}

func DartSizeof(t Term) int {
	header := t.getHeader()
	if header.DartSize >= 0 {
		return header.DartSize
	}

	fillSize(t, func(t Term) *int {
		return &t.getHeader().DartSize
	}, func(t Term) bool {
		if t, ok := t.(*Field); ok {
			return !t.SendToDart
		}
		return false
	})
	return header.DartSize
}

func FillAllSize(t Term) {
	GoSizeof(t)
	DartSizeof(t)
}
