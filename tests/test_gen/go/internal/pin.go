package internal

import dgo "github.com/hsfzxjy/dgo/go"

//dgo:export ,pinnable
type StructPinnable struct {
	dgo.PinMeta
	value int
	a     StructWithPointerToSingleType
}

//dgo:export
func (s *StructPinnable) ReturnsPinned() dgo.PinToken[StructPinnable] {
	s.Pin()
	return s.NewToken()
}

//dgo:export
func (s *StructPinnable) ReturnsMultiplePinned() (result []dgo.PinToken[StructPinnable]) {
	s.Pin()
	result = append(result, s.NewToken())
	return
}

//dgo:export
func (s *StructPinnable) AcceptToken(token dgo.PinToken[StructPinnable]) (result []dgo.PinToken[StructPinnable]) {
	result = append(result, token)
	return
}
