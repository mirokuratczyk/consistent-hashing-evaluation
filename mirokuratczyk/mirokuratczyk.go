package mirokuratczyk

import (
	"github.com/cespare/xxhash"
	"github.com/mirokuratczyk/consistent"
	"github.com/mirokuratczyk/consistent-hashing-evaluation/v2/intf"
)

type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

type mk struct {
	c *consistent.Consistent
	intf.Consistent
}

func (m *mk) Add(member intf.Member) {
	m.c.Add(member)
}

func (m *mk) LocateKey(key []byte) intf.Member {
	r := m.c.LocateKey(key)
	return intf.Member(r.String())
}

func NewConsistent() *mk {
	return &mk{
		c: consistent.New(
			nil,
			consistent.Config{
				Hasher: hasher{},
			}),
	}
}
