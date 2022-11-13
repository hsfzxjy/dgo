package dgo

type CallbackFlag uint64

const (
	cfBitsStart     = 32
	cfBitsFastStart = cfBitsStart + 4

	CF CallbackFlag = 0

	CF_POP CallbackFlag = 1 << (cfBitsStart + 0)

	CF_WITHCODE  CallbackFlag = 1 << (cfBitsStart + 1)
	CF_PACKARRAY CallbackFlag = 1 << (cfBitsStart + 2)

	CF_FAST      CallbackFlag = 1 << (cfBitsStart + 3)
	CF_FAST_VOID CallbackFlag = CF_FAST + (0 << (cfBitsStart + 4))
	CF_FAST_NIL  CallbackFlag = CF_FAST + (1 << (cfBitsStart + 4))
	CF_FAST_YES  CallbackFlag = CF_FAST + (2 << (cfBitsStart + 4))
	CF_FAST_NO   CallbackFlag = CF_FAST + (3 << (cfBitsStart + 4))

	CF_FALLIBLE CallbackFlag = 1 << (cfBitsStart + 6)

	cf_fut_reject  CallbackFlag = 0 << (cfBitsStart + 7)
	cf_fut_resolve CallbackFlag = 1 << (cfBitsStart + 7)

	cf_method_call CallbackFlag = 1 << (cfBitsStart + 8)
)

// 8 <= n <= 15
func CF_CUSTOM(n int) CallbackFlag {
	return CallbackFlag(1 << (cfBitsStart + n))
}

func (cf CallbackFlag) Pop() CallbackFlag { return cf | CF_POP }
func (cf CallbackFlag) HasPop() bool      { return cf&CF_POP != 0 }

func (cf CallbackFlag) WithCode() CallbackFlag { return cf | CF_WITHCODE }
func (cf CallbackFlag) HasWithCode() bool      { return cf&CF_WITHCODE != 0 }

func (cf CallbackFlag) PackArray() CallbackFlag { return cf | CF_PACKARRAY }
func (cf CallbackFlag) HasPackArray() bool      { return cf&CF_PACKARRAY != 0 }

func (cf CallbackFlag) Fast(kind CFFastKind) CallbackFlag {
	switch kind {
	case CFFK_NONE:
		// clear CF_FAST
		return cf & ^CF_FAST
	case CFFK_VOID, CFFK_NIL, CFFK_YES, CFFK_NO:
		// set CF_FAST
		cf = cf | CF_FAST
		cf = cf & ^((CF_FAST << 1) | (CF_FAST << 2))
		cf = cf + (CallbackFlag(kind-1) << cfBitsFastStart)
		return cf
	default:
		return cf
	}
}
func (cf CallbackFlag) HasFast() bool { return cf&CF_FAST != 0 }

func (cf CallbackFlag) FastKind() CFFastKind {
	if cf&CF_FAST == 0 {
		return CFFK_NONE
	}
	kind := cf & ((CF_FAST << 1) | (CF_FAST << 2))
	kind >>= cfBitsFastStart
	return CFFastKind(kind + 1)
}

func (cf CallbackFlag) Fallible() CallbackFlag { return cf | CF_FALLIBLE }
func (cf CallbackFlag) HasFallible() bool      { return cf&CF_FALLIBLE != 0 }

func (cf CallbackFlag) hasMethodCall() bool { return cf&cf_method_call != 0 }

type CFFastKind int

const (
	CFFK_NONE CFFastKind = iota
	CFFK_VOID
	CFFK_NIL
	CFFK_YES
	CFFK_NO
)
