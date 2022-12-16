package keepalive

import (
	"reflect"
	"sync"
	"unsafe"
)

type node struct {
	data unsafe.Pointer
	next *node
}

var pool = sync.Pool{
	New: func() any { return new(node) },
}

type Holder struct {
	head *node
}

func (h *Holder) Add(data unsafe.Pointer) {
	node := pool.Get().(*node)
	node.data = data
	node.next = h.head
	h.head = node
}

func (h *Holder) AddString(s string) {
	node := pool.Get().(*node)
	sheader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	node.data = unsafe.Pointer(sheader.Data)
	node.next = h.head
	h.head = node
}

func (h Holder) Free() {
	n := h.head
	for n != nil {
		p := n.next
		n.data = nil
		n.next = nil
		pool.Put(n)
		n = p
	}
}
