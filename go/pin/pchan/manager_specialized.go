package pchan

import (
	"sync"

	"github.com/hsfzxjy/dgo/go/pin/bitset"
)

type manager1 struct {
	heads   [1]*listener
	lidsets [1]bitset.Bitset64
}

func (c *manager1) as_chanman(n uint8) Manager {
	return Manager{
		backing: c,
		heads:   c.heads[:n],
		lidsets: c.lidsets[:n],
	}
}

func (c *manager1) recycle() {
	for i := range c.heads {
		c.heads[i].free()
		c.heads[i] = nil
	}
	c.lidsets = [1]bitset.Bitset64{}
}

type manager2 struct {
	heads   [2]*listener
	lidsets [2]bitset.Bitset64
}

func (c *manager2) as_chanman(n uint8) Manager {
	return Manager{
		backing: c,
		heads:   c.heads[:n],
		lidsets: c.lidsets[:n],
	}
}

func (c *manager2) recycle() {
	for i := range c.heads {
		c.heads[i].free()
		c.heads[i] = nil
	}
	c.lidsets = [2]bitset.Bitset64{}
}

type manager4 struct {
	heads   [4]*listener
	lidsets [4]bitset.Bitset64
}

func (c *manager4) as_chanman(n uint8) Manager {
	return Manager{
		backing: c,
		heads:   c.heads[:n],
		lidsets: c.lidsets[:n],
	}
}

func (c *manager4) recycle() {
	for i := range c.heads {
		c.heads[i].free()
		c.heads[i] = nil
	}
	c.lidsets = [4]bitset.Bitset64{}
}

type manager8 struct {
	heads   [8]*listener
	lidsets [8]bitset.Bitset64
}

func (c *manager8) as_chanman(n uint8) Manager {
	return Manager{
		backing: c,
		heads:   c.heads[:n],
		lidsets: c.lidsets[:n],
	}
}

func (c *manager8) recycle() {
	for i := range c.heads {
		c.heads[i].free()
		c.heads[i] = nil
	}
	c.lidsets = [8]bitset.Bitset64{}
}

type manager16 struct {
	heads   [16]*listener
	lidsets [16]bitset.Bitset64
}

func (c *manager16) as_chanman(n uint8) Manager {
	return Manager{
		backing: c,
		heads:   c.heads[:n],
		lidsets: c.lidsets[:n],
	}
}

func (c *manager16) recycle() {
	for i := range c.heads {
		c.heads[i].free()
		c.heads[i] = nil
	}
	c.lidsets = [16]bitset.Bitset64{}
}

type manager32 struct {
	heads   [32]*listener
	lidsets [32]bitset.Bitset64
}

func (c *manager32) as_chanman(n uint8) Manager {
	return Manager{
		backing: c,
		heads:   c.heads[:n],
		lidsets: c.lidsets[:n],
	}
}

func (c *manager32) recycle() {
	for i := range c.heads {
		c.heads[i].free()
		c.heads[i] = nil
	}
	c.lidsets = [32]bitset.Bitset64{}
}

type manager64 struct {
	heads   [64]*listener
	lidsets [64]bitset.Bitset64
}

func (c *manager64) as_chanman(n uint8) Manager {
	return Manager{
		backing: c,
		heads:   c.heads[:n],
		lidsets: c.lidsets[:n],
	}
}

func (c *manager64) recycle() {
	for i := range c.heads {
		c.heads[i].free()
		c.heads[i] = nil
	}
	c.lidsets = [64]bitset.Bitset64{}
}

var managerPools [7]sync.Pool

func init() {
	managerPools[0].New = func() any { return new(manager1) }
	managerPools[1].New = func() any { return new(manager2) }
	managerPools[2].New = func() any { return new(manager4) }
	managerPools[3].New = func() any { return new(manager8) }
	managerPools[4].New = func() any { return new(manager16) }
	managerPools[5].New = func() any { return new(manager32) }
	managerPools[6].New = func() any { return new(manager64) }
}

type imanager interface {
	as_chanman(n uint8) Manager
	recycle()
}
