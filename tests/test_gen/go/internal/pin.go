package internal

import "github.com/hsfzxjy/dgo/go/pin"

//dgo:export ,pinnable
type StructPinnable struct {
	pin.Meta
	value          int
	a              StructWithPointerToSingleType
	state          chan int
	state2         chan int
	statedyn       chan []int
	stateMemo      chan int `dgo:",memo"`
	stateBroadcast chan int `dgo:",broadcast"`
	stateBlock     chan int `dgo:",block"`
}

//dgo:export
func (s *StructPinnable) ReturnsPinned() pin.Token[StructPinnable] {
	s.Pin()
	return s.NewToken()
}

//dgo:export
func (s *StructPinnable) ReturnsMultiplePinned() (result []pin.Token[StructPinnable]) {
	s.Pin()
	result = append(result, s.NewToken())
	return
}

//dgo:export
func (s *StructPinnable) AcceptToken(token pin.Token[StructPinnable]) (result []pin.Token[StructPinnable]) {
	result = append(result, token)
	return
}
