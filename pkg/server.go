package pkg

import (
	"BrunoCoin/pkg/address"
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/peer"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"time"
)

// Checks to see that requesting node is a peer and updates last seen for the peer
func (n *Node) peerCheck(addr string) error {
	if n.PeerDb.Get(addr) == nil {
		return errors.New("request from non-peered node")
	}
	err := n.PeerDb.UpdateLastSeen(addr, uint32(time.Now().UnixNano()))
	if err != nil {
		fmt.Printf("ERROR {Node.peerCheck}: error" +
			"when calling updatelastseen.\n")
	}
	return nil
}

// Handles version request (a request to become a peer)
func (n *Node) Version(ctx context.Context, in *proto.VersionRequest) (*proto.Empty, error) {
	// Reject all outdated versions (this is not true to Satoshi Client)
	if int(in.Version) != n.Conf.Version {
		return &proto.Empty{}, nil
	}
	// If addr map is full or does not contain addr of ver, reject
	newAddr := address.New(in.AddrMe, uint32(time.Now().UnixNano()))
	if n.AddrDb.Get(newAddr.Addr) != nil {
		err := n.AddrDb.UpdateLastSeen(newAddr.Addr, newAddr.LastSeen)
		if err != nil {
			return &proto.Empty{}, nil
		}
	} else if err := n.AddrDb.Add(newAddr); err != nil {
		return &proto.Empty{}, nil
	}
	newPeer := peer.New(n.AddrDb.Get(newAddr.Addr), in.Version, in.BestHeight)
	// Check if we are waiting for a ver in response to a ver, do not respond if this is a confirmation of peering
	pendingVer := newPeer.Addr.SentVer != time.Time{} && newPeer.Addr.SentVer.Add(n.Conf.VerTimeout).After(time.Now())
	if n.PeerDb.Add(newPeer) && !pendingVer {
		newPeer.Addr.SentVer = time.Now()
		_, err := newAddr.VersionRPC(&proto.VersionRequest{
			Version:    uint32(n.Conf.Version),
			AddrYou:    in.AddrMe,
			AddrMe:     n.Addr,
			BestHeight: uint32(n.Chain.Length()),
		})
		if err != nil {
			return &proto.Empty{}, err
		}
	}
	return &proto.Empty{}, nil
}

// Handles get blocks request (request for blocks past a certain block)
func (n *Node) GetBlocks(ctx context.Context, in *proto.GetBlocksRequest) (*proto.GetBlocksResponse, error) {
	blockHashes := make([]string, 0)
	if ind := n.Chain.IndexOf(in.TopBlockHash); ind != -1 && ind < n.Chain.Length() {
		upperIndex := n.Chain.Length()
		// Can send a maximum of 50 0 headers
		if ind+500 < upperIndex {
			upperIndex = ind + 500
		}
		for _, bn := range n.Chain.Slice(ind+1, upperIndex) {
			blockHashes = append(blockHashes, bn.Hash())
		}
	}
	return &proto.GetBlocksResponse{BlockHashes: blockHashes}, nil
}

// Handles get data request (request for a specific block identified by its hash)
func (n *Node) GetData(ctx context.Context, in *proto.GetDataRequest) (*proto.GetDataResponse, error) {
	blk := n.Chain.Get(in.BlockHash)
	if blk == nil {
		utils.Debug.Printf("Node {%v} received a data req from the network for a block {%v} that could not be found locally.\n",
			n.Addr, in.BlockHash)
		return &proto.GetDataResponse{}, nil
	}
	return &proto.GetDataResponse{Block: blk.Serialize()}, nil
}

// Handles send addresses request (request for nodes to peer with the requesting node)
func (n *Node) SendAddresses(ctx context.Context, in *proto.Addresses) (*proto.Empty, error) {
	// Forward nodes to all neighbors if new nodes were found (without redundancy)
	foundNew := false
	for _, addr := range in.Addrs {
		if addr.Addr == n.Addr {
			continue
		}
		newAddr := address.New(addr.Addr, addr.LastSeen)
		if p := n.PeerDb.Get(addr.Addr); p != nil {
			if p.Addr.LastSeen < addr.LastSeen {
				err := n.PeerDb.UpdateLastSeen(addr.Addr, addr.LastSeen)
				if err != nil {
					fmt.Printf("ERROR {Node.SendAddresses}: error" +
						"when calling updatelastseen.\n")
				}
				foundNew = true
			}
		} else if a := n.AddrDb.Get(addr.Addr); a != nil {
			if a.LastSeen < addr.LastSeen {
				err := n.AddrDb.UpdateLastSeen(addr.Addr, addr.LastSeen)
				if err != nil {
					fmt.Printf("ERROR {Node.SendAddresses}: error" +
						"when calling updatelastseen.\n")
				}
			}
		} else {
			err := n.AddrDb.Add(newAddr)
			if err == nil {
				foundNew = true
			}
		}
		// Try to connect to each new address as true peers (it is okay if this is repeated, this may be a reboot)
		go func() {
			_, err := newAddr.VersionRPC(&proto.VersionRequest{
				Version:    uint32(n.Conf.Version),
				AddrYou:    newAddr.Addr,
				AddrMe:     n.Addr,
				BestHeight: uint32(n.Chain.Length()),
			})
			if err != nil {
				utils.Debug.Printf("%v recieved no response from VersionRPC to %v",
					utils.FmtAddr(n.Addr), utils.FmtAddr(addr.Addr))
			}
		}()
	}
	if foundNew {
		bcPeers := n.PeerDb.GetRandom(2, []string{n.Addr})
		for _, p := range bcPeers {
			_, err := p.Addr.SendAddressesRPC(in)
			if err != nil {
				utils.Debug.Printf("%v recieved no response from SendAddressesRPC to %v",
					utils.FmtAddr(n.Addr), utils.FmtAddr(p.Addr.Addr))
			}
		}
	}
	return &proto.Empty{}, nil
}

// Handles get addresses request (request for all known addresses from a specific node)
func (n *Node) GetAddresses(ctx context.Context, in *proto.Empty) (*proto.Addresses, error) {
	utils.Debug.Printf("Node {%v} received a GetAddresses req from the network.\n",
		n.Addr)
	return &proto.Addresses{Addrs: n.AddrDb.Serialize()}, nil
}

// Handles forward transaction request (tx propagation)
func (n *Node) ForwardTransaction(ctx context.Context, in *proto.Transaction) (*proto.Empty, error) {
	t := tx.Deserialize(in)
	if n.TxMap[t.Hash()] {
		return &proto.Empty{}, nil
	}
	if !n.ChkTx(t) {
		utils.Debug.Printf("%v recieved invalid %v", utils.FmtAddr(n.Addr), t.NameTag())
		return &proto.Empty{}, errors.New("transaction is not valid")
	}
	utils.Debug.Printf("%v recieved valid %v", utils.FmtAddr(n.Addr), t.NameTag())
	if n.Conf.MnrConf.HasMnr {
		go n.Mnr.HndlTx(t)
	}
	n.TxMap[t.Hash()] = true
	for _, p := range n.PeerDb.List() {
		go func(addr *address.Address) {
			_, err := addr.ForwardTransactionRPC(t.Serialize())
			if err != nil {
				utils.Debug.Printf("%v recieved no response from ForwardTransaction to %v",
					utils.FmtAddr(n.Addr), utils.FmtAddr(p.Addr.Addr))
			}
		}(p.Addr)
	}
	return &proto.Empty{}, nil
}

// Handles forward block request (block propagation)
func (n *Node) ForwardBlock(ctx context.Context, in *proto.Block) (*proto.Empty, error) {
	b := block.Deserialize(in)
	// Ignore if already seen block
	n.BlockMapMutex.Lock()
	if n.BlockMap[b.Hash()] {
		n.BlockMapMutex.Unlock()
		return &proto.Empty{}, nil
	}
	n.BlockMap[b.Hash()] = true
	n.BlockMapMutex.Unlock()
	if !n.ChkBlk(b) {
		utils.Debug.Printf("%v recieved invalid %v", utils.FmtAddr(n.Addr), b.NameTag())
		return &proto.Empty{}, errors.New("block is not valid")
	}
	mnChn := n.Chain.IsEndMainChain(b)
	n.Chain.Add(b)
	if n.Conf.MnrConf.HasMnr && mnChn {
		go n.Mnr.HndlBlk(b)
	}
	if n.Conf.WtConf.HasWt && mnChn {
		blks := n.Chain.Slice(n.Chain.Length()-n.Conf.WtConf.SafeBlkAmt, n.Chain.Length())
		if len(blks) == n.Conf.WtConf.SafeBlkAmt {
			go n.Wallet.HndlBlk(blks[0])
		}
	}
	for _, p := range n.PeerDb.List() {
		go func(addr *address.Address) {
			_, err := addr.ForwardBlockRPC(b.Serialize())
			if err != nil {
				utils.Debug.Printf("%v recieved no response from ForwardBlockRPC to %v",
					utils.FmtAddr(n.Addr), utils.FmtAddr(p.Addr.Addr))
			}
		}(p.Addr)
	}
	return &proto.Empty{}, nil
}
