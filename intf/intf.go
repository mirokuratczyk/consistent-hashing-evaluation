package intf

type Member string

func (m Member) String() string {
	return string(m)
}

type Consistent interface {
	Add(member Member)
	LocateKey(key []byte) Member
}
