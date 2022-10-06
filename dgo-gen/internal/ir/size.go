package ir

func fillSize(t Term) {
	cntSize := 0
	t.Traverse(func(t Term) {
		t.getHeader().Size = cntSize
		if _, ok := t.(*Basic); ok {
			cntSize++
		}
	}, func(t Term) {
		header := t.getHeader()
		base := header.Size
		if t, ok := t.(*Array); ok {
			arrSize := (cntSize - base) * int(t.Len)
			cntSize = base + arrSize
			header.Size = arrSize + 1
		} else {
			header.Size = cntSize - base + 1
		}
	})
}

func Sizeof(t Term) int {
	if _, ok := t.(*Basic); ok {
		return 1
	}
	header := t.getHeader()
	if header.Size > 0 {
		return int(header.Size) - 1
	}

	fillSize(t)
	return header.Size
}
