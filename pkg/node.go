package pkg

import (
	"BrunoCoin/pkg/address"
	"BrunoCoin/pkg/address/addressdb"
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/id"
	"BrunoCoin/pkg/miner"
	"BrunoCoin/pkg/peer"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"BrunoCoin/pkg/wallet"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// Node is the interface for interacting with
// the cryptocurrency. The node handles all top
// level logic and communication between different
// pieces of functionality. For example, it handles
// the logic of maintaining a gRPC server and
// passing transactions and blocks to the miner,
// wallet, chain, and other nodes on the network.
// It is also the interface between the person using
// the computer which means all transaction requests
// and directives to stop or resume the node is done
// on the node object.
// *proto.UnimplementedBrunoCoinServer
// Server *grpc.Server
// Conf *Config the settings for the node
// Addr string the address that the node is listening
// to traffic on
// Id   id.ID the id of the node
// Chain  *blockchain.Blockchain the blockchain
// Wallet *wallet.Wallet the wallet
// Mnr    *miner.Miner the miner
// fGetAddr bool
// AddrDb   addressdb.AddressDb a database of addresses
// of nodes that it knows about in the network
// PeerDb   peer.PeerDb a database of peers the node
// is currently connected to
// TxMap    map[string]bool a map used to keep track
// of whether a transaction has been seen on the network
// before or not
// BlockMap map[string]bool a map used to keep track
// of whether a block has been seen on the network
// before or not
// Paused bool
type Node struct {
	*proto.UnimplementedBrunoCoinServer
	Server *grpc.Server

	Conf *Config
	Addr string
	Id   id.ID

	Chain  *blockchain.Blockchain
	Wallet *wallet.Wallet
	Mnr    *miner.Miner

	fGetAddr bool // starts false, set to true when we request addresses from a node, cleared when we receive less than 1000 addresses from a node

	AddrDb        addressdb.AddressDb
	PeerDb        peer.PeerDb
	TxMap         map[string]bool
	BlockMap      map[string]bool
	BlockMapMutex sync.Mutex

	Paused bool
}

// SendTx (SendTransaction) sends a transaction to
// someone (identified by their public key) with a certain
// amount of money and a certain fee. The fee is used to
// incentivize miners to mine their transaction to the
// blockchain.
// Inputs:
// amt uint32 the amount of money to be paid to someone
// fee uint32 the amount of extra money to be paid to the
// miner who mines your transaction
// pubK []byte the public key of the person you are sending
// money to
func (n *Node) SendTx(amt uint32, fee uint32, pubK []byte) {
	if amt <= 0 {
		utils.Debug.Printf("Node {%v} received non-positive amount", n)
		return
	}
	if pubK == nil {
		utils.Debug.Printf("Node {%v} received nil pubK", n)
		return
	}
	txR := &wallet.TxReq{
		PubK: pubK,
		Amt:  amt,
		Fee:  fee,
	}
	go n.Wallet.HndlTxReq(txR)
}

// New returns a new Node object based on
// a configuration
// Inputs:
// conf *Config the desired configuration
// of the Node
// Returns:
// *Node a pointer to the new node object
func New(conf *Config) *Node {
	n := &Node{Conf: conf}
	if conf.CstmID {
		n.Id = conf.CstmIDObj
	} else {
		n.Id, _ = id.New(n.Conf.IdConf)
	}
	n.Chain = blockchain.New(n.Conf.ChainConf)
	n.Wallet = wallet.New(n.Conf.WtConf, n.Id, n.Chain)
	n.Mnr = miner.New(n.Conf.MnrConf, n.Id)

	n.AddrDb = addressdb.New(true, 1000)
	n.PeerDb = peer.NewDb(true, 200, "")
	n.TxMap = make(map[string]bool)
	n.BlockMap = make(map[string]bool)

	return n
}

// Start starts a node on the network. At first, the node is
// not technically connected to the network, since it has no
// one to connect to. So, this method opens up a listener and
// creates a gRPC server that it can used to make and listen to
// requests on the network. It also starts another go routine
// for listening to messages from the wallet and/or the miner.
func (n *Node) Start() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addr := fmt.Sprintf("%v:%v", hostname, n.Conf.Port)
	n.Addr = addr
	n.PeerDb.SetAddr(addr)
	utils.Debug.Printf("%v started", utils.FmtAddr(n.Addr))
	if n.Conf.MnrConf.HasMnr {
		n.Mnr.SetAddr(addr)
	}
	if n.Conf.ChainConf.HasChn {
		n.Chain.SetAddr(addr)
	}
	if n.Conf.WtConf.HasWt {
		n.Wallet.SetAddr(addr)
	}
	n.StartServer(addr)
	go func() {
		if n.Conf.MnrConf.HasMnr {
			for {
				select {
				case t := <-n.Wallet.SendTx:
					n.HndlWtTx(t)
				case b := <-n.Mnr.SendBlk:
					n.HndlMnrBlk(b)
				}
			}
		} else {
			for {
				select {
				case t := <-n.Wallet.SendTx:
					n.HndlWtTx(t)
				}
			}
		}
	}()
}

// HndlMnrBlk (HandleMinerBlock) handles a block
// that was just made by the miner. It does this
// by sending the block to the chain so it can be
// added, to the wallet, and to the network to be
// broadcast. It is also added to the map of
// seen blocks.
// Inputs:
// b *block.Block the block mined by the miner
func (n *Node) HndlMnrBlk(b *block.Block) {
	n.BlockMapMutex.Lock()
	n.BlockMap[b.Hash()] = true
	n.BlockMapMutex.Unlock()
	n.Chain.Add(b)
	if n.Conf.WtConf.HasWt {
		blks := n.Chain.Slice(n.Chain.Length()-n.Conf.WtConf.SafeBlkAmt, n.Chain.Length())
		if len(blks) == n.Conf.WtConf.SafeBlkAmt {
			go n.Wallet.HndlBlk(blks[0])
		}
	}
	for _, p := range n.PeerDb.List() {
		utils.Debug.Printf("%v sending %v to %v", utils.FmtAddr(n.Addr), b.NameTag(), utils.FmtAddr(p.Addr.Addr))
		go func(addr *address.Address) {
			_, err := addr.ForwardBlockRPC(b.Serialize())
			if err != nil {
				utils.Debug.Printf("%v recieved no response from ForwardBlockRPC to %v",
					utils.FmtAddr(n.Addr), utils.FmtAddr(p.Addr.Addr))
			}
		}(p.Addr)
	}
}

// GetBalance returns the balance (amount of money)
// that someone currently has.
// Inputs:
// pk string the public key of the person that the
// balance wants to be known for.
// Returns:
// uint32 the amount of money (the balance) that
// the person with that public key has
func (n *Node) GetBalance(pk string) uint32 {
	return n.Chain.GetBalance(pk)
}

// HndlWtTx (HandleWalletTransaction) handles a new
// transaction being created by the wallet. It does
// this by sending that transaction to the network as
// well as to the miner. Lastly, it is also added
// to the map of seen transactions.
// Inputs:
// t *tx.Transaction the transaction that was just
// made by the wallet.
func (n *Node) HndlWtTx(t *tx.Transaction) {
	if n.Conf.MnrConf.HasMnr {
		go n.Mnr.HndlTx(t)
	}
	n.TxMap[t.Hash()] = true
	for _, p := range n.PeerDb.List() {
		d := t.Serialize()
		utils.Debug.Printf("%v sending %v to %v", utils.FmtAddr(n.Addr), t.NameTag(), utils.FmtAddr(p.Addr.Addr))
		go func(addr *address.Address) {
			_, err := addr.ForwardTransactionRPC(d)
			if err != nil {
				utils.Debug.Printf("%v recieved no response from ForwardTransactionRPC to %v",
					utils.FmtAddr(n.Addr), utils.FmtAddr(p.Addr.Addr))
			}
		}(p.Addr)
	}
}

// StartMiner starts the miner, which means the miner
// is now actively waiting for enough transactions
// to mine.
func (n *Node) StartMiner() {
	n.Mnr.StartMiner()
}

// This connects to a certain peer in the network. This just
// serves as an interface for the real functionality contained
// within the Router.
// Inputs:
// addr string the address of the node that you want
// to connect to.
func (n *Node) ConnectToPeer(addr string) {
	a := address.New(addr, 0)
	_, err := a.VersionRPC(&proto.VersionRequest{
		Version:    uint32(n.Conf.Version),
		AddrYou:    addr,
		AddrMe:     n.Addr,
		BestHeight: uint32(n.Chain.Length()),
	})
	if err != nil {
		utils.Debug.Printf("%v recieved no response from VersionRPC to %v",
			utils.FmtAddr(n.Addr), utils.FmtAddr(addr))
	}
}

// BroadcastAddr
func (n *Node) BroadcastAddr() {
	myAddr := proto.Address{Addr: n.Addr, LastSeen: uint32(time.Now().UnixNano())}
	for _, p := range n.PeerDb.List() {
		go func(addr *address.Address) {
			_, err := addr.SendAddressesRPC(&proto.Addresses{Addrs: []*proto.Address{&myAddr}})
			if err != nil {
				utils.Debug.Printf("%v recieved no response from SendAddressesRPC to %v",
					utils.FmtAddr(n.Addr), utils.FmtAddr(p.Addr.Addr))
			}
		}(p.Addr)
	}
}

// Bootstrap attempts to build a blockchain based on the
// pre-existing one that other nodes have. This may happen
// when a node first joins the network, or if the node left
// the network for a while (paused), then rejoined.
func (n *Node) Bootstrap() error {
	utils.Debug.Printf("%v bootstrapping from %v peers with top block %v", utils.FmtAddr(n.Addr), len(n.PeerDb.List()), n.Chain.LastBlock.NameTag())
	topBlockHash := n.Chain.GetLastBlock().Hash()
	var wg sync.WaitGroup
	var longestRes *proto.GetBlocksResponse
	var addr *address.Address
	if len(n.PeerDb.List()) == 0 {
		return errors.New("no peers to bootstrap from")
	}
	for _, p := range n.PeerDb.List() {
		wg.Add(1)
		go func(p *peer.Peer) {
			res, err := p.Addr.GetBlocksRPC(&proto.GetBlocksRequest{TopBlockHash: topBlockHash})
			if err != nil {
				wg.Done()
				return
			}
			if longestRes == nil || len(res.BlockHashes) > len(longestRes.BlockHashes) {
				longestRes = res
				addr = p.Addr
			}
			wg.Done()
		}(p)
	}
	wg.Wait()
	if longestRes == nil {
		return errors.New("no peers gave responses")
	}
	chkOrf := len(longestRes.BlockHashes) <= 2
	for _, h := range longestRes.BlockHashes {
		pb, _ := addr.GetDataRPC(&proto.GetDataRequest{BlockHash: h})
		b := block.Deserialize(pb.Block)
		n.BlockMapMutex.Lock()
		n.BlockMap[b.Hash()] = true
		n.BlockMapMutex.Unlock()
		n.Chain.Add(b)
		if chkOrf {
			n.Mnr.HndlChkBlk(b)
		}
	}
	return nil
}

func (n *Node) StartServer(addr string) {
	lis, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}
	// Open node to connections
	n.Server = grpc.NewServer()
	proto.RegisterBrunoCoinServer(n.Server, n)
	go func() {
		err := n.Server.Serve(lis)
		if err != nil {
			fmt.Printf("ERROR {Node.StartServer}: error" +
				"when trying to serve server")
		}
	}()
}

func (n *Node) PauseNetwork() {
	n.Server.Stop()
	utils.Debug.Printf("%v paused", utils.FmtAddr(n.Addr))
}

func (n *Node) ResumeNetwork() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addr := fmt.Sprintf("%v:%v", hostname, n.Conf.Port)
	n.StartServer(addr)
	utils.Debug.Printf("%v resumed", utils.FmtAddr(n.Addr))
}

// This kills any threads currently managed by the Node or that
// it previously started. It also does any necessary clean up.
func (n *Node) Kill() {
	n.Server.GracefulStop()
}
