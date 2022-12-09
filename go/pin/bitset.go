package pin

import (
	"math"
	"sync"

	"github.com/bits-and-blooms/bitset"
)

var bitsetPool sync.Pool

func init() {
	bitsetPool.New = func() any { return bitset.New(0) }
}

func bitsetRecycle(s *bitset.BitSet) {
	bitsetPool.Put(s)
}

func bitsetGet() *bitset.BitSet {
	return bitsetPool.Get().(*bitset.BitSet)
}

func bitsetFillFirstClear(s *bitset.BitSet) (uint16, bool) {
	av, ok := s.NextClear(0)
	if !ok {
		av = s.Count()
	}
	// limit the index in range [0, MAX_UINT16-1], so that the total number
	// of bits set can be represented by uint16
	if av >= math.MaxUint16 {
		return 0, false
	}
	s.Set(av)
	return uint16(av), true
}
