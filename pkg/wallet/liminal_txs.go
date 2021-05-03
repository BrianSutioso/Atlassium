package wallet

import (
	"BrunoCoin/pkg/block/tx"
	"sync"
)

// LiminalTxs (LiminalTransactions) are
// transactions that have been made by the
// wallet but are not considered valid yet.
// This is because the transactions either
// have not been mined to a block or the
// block doesn't have enough POW on top of it.
// TxQ is a maximum priority queue that stores
// transactions with priority being equivalent
// to how many blocks have been seen since the
// transaction was initially made.
// TxRplyThresh is the threshold priority for
// having blocks removed from the TxQ, sent out
// again, and added back with a priority of 0.
type LiminalTxs struct {
	TxQ          *tx.Heap
	TxRplyThresh uint32
	mutex        sync.Mutex
}

// NewLmnlTxs (NewLiminalTransactions) returns
// a new Liminal Transactions object.
// Inputs:
// c *Config the configuration for the wallet
func NewLmnlTxs(c *Config) *LiminalTxs {
	return &LiminalTxs{
		TxQ:          tx.NewTxHeap(),
		TxRplyThresh: c.TxRplyThresh,
	}
}

// ChkTxs (CheckTransactions) checks that the inputted
// transactions from the new block aren't the same. This
// new block is assumed to have enough POW of work on top
// of it. If any transactions are the same, they are removed.
// Otherwise, the priorities are incremented, and any
// transaction with too large of a priority needs to be
// returned so it can be sent out again.
// Inputs:
// txs []*tx.Transaction a list of transactions that were in
// a valid block
// Returns:
//[]*tx.Transaction transactions with priorities above
// l.TxRplyThresh that are removed
//[]*tx.Transaction transactions from the new block that are already
//in LiminalTxs, so removed from LiminalTxs bc duplicates
func (l *LiminalTxs) ChkTxs(txs []*tx.Transaction) ([]*tx.Transaction, []*tx.Transaction) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.TxQ.IncAll()
	duplicates := l.TxQ.Rmv(txs)
	abvThreshold := l.TxQ.RemAbv(l.TxRplyThresh)

	return abvThreshold, duplicates
}

// Add adds a transaction to the liminal transactions.
// It is basically a wrapper around the heap add. The
// priority is 0, since the transaction was just made
// and no blocks have been retrieved since.
// Inputs:
// t *tx.Transaction the transaction to be added
func (l *LiminalTxs) Add(t *tx.Transaction) {
	l.mutex.Lock()

	l.TxQ.Add(0, t)

	l.mutex.Unlock()
}
