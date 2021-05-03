package peer

type PeerDb interface {
	Add(*Peer) bool
	Get(string) *Peer
	UpdateLastSeen(string, uint32) error
	List() []*Peer
	GetRandom(int, []string) []*Peer
	In(string) bool
	SetAddr(string)
}

func NewDb(eph bool, limit int, addr string) PeerDb {
	return &EphemeralPeerDb{peers: make(map[string]*Peer), limit: limit, Addr: addr}
}
