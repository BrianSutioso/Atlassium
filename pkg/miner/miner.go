package miner

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/id"
	"BrunoCoin/pkg/utils"
	"fmt"
	"sync"

	"go.uber.org/atomic"
)

// Miner supports the functionality of mining new transactions broadcast from the network to a new block.
// Conf represents the configuration (settings) for the miner.
// Id represents the identity of the miner, so that the miner can properly make the coinbase transaction.
// TxP contains all transactions that the miner is either waiting to mine, or is mining.
// MiningPool contains all transactions that the miner is currently mining.
// PrvHsh represents the hash of the last block on the main chain.
// ChnLen is the length of the main chain.
// Active is a channel used to entirely shut down the miner's ability to mine.
// Mining tells whether the miner is currently mining.
// SendBlk is used to send newly mined blocks to the node in order to be broadcast on the network.
// PoolUpdated is used to send alerts of pool updates to the miner
type Miner struct {
	Conf *Config
	Id   id.ID

	TxP        *TxPool
	MiningPool MiningPool

	PrvHsh string
	Addr   string
	ChnLen *atomic.Uint32

	Active *atomic.Bool
	Mining *atomic.Bool

	SendBlk     chan *block.Block
	PoolUpdated chan bool

	mutex sync.Mutex
}

// New constructs a new Miner according to a config and the id of a node.
func New(c *Config, id id.ID) *Miner {
	if !c.HasMnr {
		return nil
	}
	return &Miner{
		Conf:        c,
		Id:          id,
		TxP:         NewTxPool(c),
		MiningPool:  []*tx.Transaction{},
		PrvHsh:      blockchain.GenesisBlock(blockchain.DefaultConfig()).Hash(),
		ChnLen:      atomic.NewUint32(1),
		SendBlk:     make(chan *block.Block),
		PoolUpdated: make(chan bool),
		Mining:      atomic.NewBool(false),
		Active:      atomic.NewBool(false),
	}
}

// SetAddr (SetAddress) sets the address of the node that the miner is currently on.
func (m *Miner) SetAddr(a string) {
	m.mutex.Lock()
	m.Addr = a
	m.mutex.Unlock()
}

// StartMiner is a wrapper around the mine method just in case any additional work is needed to do before or after
// mining in the future.
func (m *Miner) StartMiner() {
	m.Active.Store(true)
	go m.Mine()
	m.PoolUpdated <- true
}

// HndlBlk (HandleBlock) handles a validated block from the network. The transactions on the block need to be checked
// with the transaction pool, in case the transaction pool has any transactions that have already been mined. This has
// to be done with the orphan pool as well, except, the orphans have to be recategorized. The orphans that are no
// longer orphans have to be added to the transaction pool. Also, the miner's perspective of the hash of the last block
// on the main chain needs to be reset, and the chain length needs to be updated. Lastly, the miner needs to restart.
// Inputs:
// b - a new block that is being added to the blockchain.
func (m *Miner) HndlBlk(b *block.Block) {
	if b != nil {
		m.SetHash(b.Hash())
		m.IncChnLen()
		m.HndlChkBlk(b)
	}
	return
}

// HndlChkBlk (HandleCheckBlock) handles updating
// the transaction pool and the orphan pool based
// on the new transactions in the block.
func (m *Miner) HndlChkBlk(b *block.Block) {
	if b == nil {
		fmt.Println("ERROR {m.HndlChkBlk}: received a nil block")
		return
	}

	m.TxP.ChkTxs(b.Transactions)

	if m.Active.Load() {
		m.PoolUpdated <- true
	}

	return
}

// HndlTx (HandleTransaction) handles a validated transaction from the network. If the transaction is not an orphan, it
// is added to the transaction pool. If the miner isn't currently mining and the priority threshold is met, then the
// miner is told to mine. If the transaction is an orphan, then it is added to the orphan pool.
// Inputs:
// t *tx.Transaction the validated transaction that was received from the network
func (m *Miner) HndlTx(t *tx.Transaction) {
	if t == nil {
		fmt.Println("ERROR {m.HndlTx}: received a nil transaction")
		return
	}

	m.TxP.Add(t)

	if m.Active.Load() {
		m.PoolUpdated <- true
	}

	return
}

// SetChnLen (SetChainLength) sets the miner's perspective of the length of the main chain.
// Inputs:
// l - the most updated length of the blockchain so that the miner can appropriately calculate its minting reward
func (m *Miner) SetChnLen(l uint32) {
	m.ChnLen.Store(l)
}

// SetHash sets the previous hash of the block the miner is trying to append to.
// Inputs:
// h - the hash of the new previous block that the miner is trying to append to. Represented as a hex string.
func (m *Miner) SetHash(h string) {
	m.mutex.Lock()
	m.PrvHsh = h
	m.mutex.Unlock()
}

// IncChnLen (IncrementChainLength) increments the miner's perspective of the length of the main chain.
func (m *Miner) IncChnLen() {
	m.ChnLen.Inc()
}

func (m *Miner) Pause() {
	m.Active.Store(false)
	m.PoolUpdated <- true
	utils.Debug.Printf("%v paused mining", utils.FmtAddr(m.Addr))
}

func (m *Miner) Resume() {
	m.Active.Store(true)
	m.PoolUpdated <- true
	utils.Debug.Printf("%v resumed mining", utils.FmtAddr(m.Addr))
}

// Kill closes all of the miner's channels and stops the current mining process.
func (m *Miner) Kill() {
	m.Active.Store(false)
	close(m.PoolUpdated)
	close(m.SendBlk)
}
