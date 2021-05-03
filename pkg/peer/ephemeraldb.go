package peer

import (
	"errors"
	"math/rand"
)

type EphemeralPeerDb struct {
	peers map[string]*Peer
	limit int
	Addr string
}

func (pdb *EphemeralPeerDb) In(k string) bool {
	_, in := pdb.peers[k]
	return in
}

func (pdb *EphemeralPeerDb) SetAddr(addr string) {
	pdb.Addr = addr
}

// Returns true if peer existed already or was added
func (pdb *EphemeralPeerDb) Add(p *Peer) bool {
	oldP := pdb.peers[p.Addr.Addr]
	if (oldP != nil && p.Addr.LastSeen != oldP.Addr.LastSeen) || (oldP == nil && len(pdb.peers) < pdb.limit) {
		pdb.peers[p.Addr.Addr] = p
		//utils.Debug.Printf("%v added peer %v", utils.FmtAddr(pdb.Addr), utils.FmtAddr(p.Addr.Addr))
		return true
	}
	return false
}

func (pdb *EphemeralPeerDb) Get(addr string) *Peer {
	return pdb.peers[addr]
}

func (pdb *EphemeralPeerDb) UpdateLastSeen(addr string, lastSeen uint32) error {
	p := pdb.peers[addr]
	if p == nil {
		return errors.New("peer not found")
	}
	p.Addr.LastSeen = lastSeen
	return nil
}

// Get up to n random peers
func (pdb *EphemeralPeerDb) GetRandom(n int, exclude []string) []*Peer {
	peers := make([]*Peer, 0)
	if n >= len(pdb.peers) {
		for _, peer := range pdb.peers {
			peers = append(peers, peer)
		}
		return peers
	}
	blacklistedAddrs := make(map[string]bool)
	for _, addr := range exclude {
		blacklistedAddrs[addr] = true
	}
	keys := make([]string, 0)
	for key := range pdb.peers {
		if blacklistedAddrs[key] == false {
			keys = append(keys, key)
		}
	}
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	randKeys := keys[:n]
	for _, key := range randKeys {
		peers = append(peers, pdb.peers[key])
	}
	return peers
}

func (pdb *EphemeralPeerDb) List() []*Peer {
	peers := make([]*Peer, 0)
	for _, peer := range pdb.peers {
		peers = append(peers, peer)
	}
	return peers
}
