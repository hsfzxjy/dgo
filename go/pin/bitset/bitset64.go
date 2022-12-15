package bitset

import "math/bits"

type Bitset64 uint64

func (b Bitset64) IsEmpty() bool     { return b == 0 }
func (b Bitset64) Count() int        { return bits.OnesCount64(uint64(b)) }
func (b Bitset64) Test(i uint8) bool { return i < 64 && b&(1<<i) != 0 }
func (b *Bitset64) Set(i uint8)      { *b |= 1 << i }
func (b *Bitset64) Clear(i uint8)    { *b &^= 1 << i }

func (b *Bitset64) PickSet() (result uint8, success bool) {
	const full = ^uint64(0)
	v := uint64(*b)
	if v == full {
		return 0, false
	}
	x := bits.TrailingZeros64(^v)
	*b |= 1 << x
	return uint8(x), true
}
