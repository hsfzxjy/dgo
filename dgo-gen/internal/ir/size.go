package ir

type hasCachedSize interface {
	getCachedSize() *cachedSize
}

type cachedSize struct {
	Size int
}

func (s *cachedSize) getCachedSize() *cachedSize { return s }

func fillSize(t Term) {
	cntSize := 0
	t.Traverse(func(t Term, vk visitKind) {
		var cache *cachedSize
		if t, ok := t.(hasCachedSize); ok {
			cache = t.getCachedSize()
		}
		switch vk {
		case vkLeaf:
			cntSize += 1
		case vkEnter:
			cache.Size = cntSize
		case vkLeave:
			base := cache.Size
			if t, ok := t.(*Array); ok {
				arrSize := (cntSize - base) * int(t.Len)
				cntSize = base + arrSize
				cache.Size = arrSize + 1
			} else {
				cache.Size = cntSize - base + 1
			}
		}
	})
}

func Sizeof(t Term) int {
	if _, ok := t.(*Basic); ok {
		return 1
	}
	cached := t.(hasCachedSize).getCachedSize()
	if cached.Size > 0 {
		return int(cached.Size) - 1
	}

	fillSize(t)
	return cached.Size
}
