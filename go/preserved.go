package dgo

import _ "unsafe"

type _PreservedGoCall struct {
	kind uint64
	port *Port
}

const (
	pgcTokenDispose uint64 = iota + 1
	pgcChanListen
	pgcChanCancelListen
)

func parseUntypedToken(objs []*Dart_CObject) (version uint16, lid uint8, data uintptr) {
	version = uint16(cobjectParseInt(objs[0]))
	lid = uint8(cobjectParseInt(objs[1]))
	data = uintptr(cobjectParseInt(objs[2]))
	return
}

func (c _PreservedGoCall) specialInt() {}
func (c _PreservedGoCall) handleCObjects(objs []*Dart_CObject) {
	switch c.kind {
	case pgcTokenDispose:
		version, lid, data := parseUntypedToken(objs)
		pin_TokenDispose(version, lid, data)
	case pgcChanListen:
		version, lid, data := parseUntypedToken(objs)
		chid := uint8(cobjectParseInt(objs[3]))
		dcb := uint32(cobjectParseInt(objs[4]))
		pin_ChanListen(version, lid, data, chid, dcb, c.port)
	case pgcChanCancelListen:
		version, lid, data := parseUntypedToken(objs)
		chid := uint8(cobjectParseInt(objs[3]))
		pin_ChanCancelListen(version, lid, data, chid)
	}
}

func pin_TokenDispose(version uint16, lid uint8, data uintptr)
func pin_ChanListen(version uint16, lid uint8, data uintptr, chid uint8, dcb uint32, port *Port)
func pin_ChanCancelListen(version uint16, lid uint8, data uintptr, chid uint8)
