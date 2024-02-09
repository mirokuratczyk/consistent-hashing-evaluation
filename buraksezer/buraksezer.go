package buraksezer

import (
	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
	"github.com/mirokuratczyk/consistent-hashing-evaluation/v2/intf"
)

type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

type bs struct {
	c *consistent.Consistent
	intf.Consistent
}

func (m *bs) Add(member intf.Member) {
	m.c.Add(member)
}

func (m *bs) LocateKey(key []byte) intf.Member {
	r := m.c.LocateKey(key)
	return intf.Member(r.String())
}

func NewConsistent() *bs {

	return &bs{
		c: consistent.New(nil,
			consistent.Config{
				Hasher:            hasher{},
				PartitionCount:    consistent.DefaultPartitionCount * 4,
				ReplicationFactor: consistent.DefaultReplicationFactor,
				Load:              consistent.DefaultLoad,
			}),
	}
}
