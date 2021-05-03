package peer

import (
	"BrunoCoin/pkg/address"
)

type Peer struct {
	Addr          *address.Address
	Version       uint32
	bestHeight    uint32
}

func New(addr *address.Address, version uint32, bestHeight uint32) *Peer {
	return &Peer{Addr: addr, Version: version, bestHeight: bestHeight}
}
