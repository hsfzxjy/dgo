package dgo

import (
	"math"
)

const _canonicalNANbits uint64 = 0xFFFF_FFFF_FFFF_FFFF

func canonicalizeNAN(x float64) float64 {
	if math.IsNaN(x) {
		return math.Float64frombits(_canonicalNANbits)
	}
	return x
}
