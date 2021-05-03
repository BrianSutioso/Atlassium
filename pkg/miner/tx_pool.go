package miner

import (
	"BrunoCoin/pkg/block/tx"
	"fmt"
	"sync"

	"go.uber.org/atomic"
)

// TxPool represents all the valid transactions
// that the miner can mine.
// CurPri is the current cumulative priority of
// all the transactions.
// PriLim is the cumulative priority threshold
// needed to surpass in order to start mining.
// TxQ is the transaction maximum priority queue
// that the transactions are stored in.
// Ct is the current count of the transactions
// in the pool.
// Cap is the maximum amount of allowed
// transactions to store in the pool.
type TxPool struct {
	CurPri *atomic.Uint32
	PriLim uint32

	TxQ   *tx.Heap
	Ct    *atomic.Uint32
	Cap   uint32
	mutex sync.Mutex
}

// Length returns the count of transactions
// currently in the pool.
// Returns:
// uint32 the count (Ct) of the pool
func (tp *TxPool) Length() uint32 {
	return tp.Ct.Load()
}

// NewTxPool constructs a transaction pool.
func NewTxPool(c *Config) *TxPool {
	return &TxPool{
		CurPri: atomic.NewUint32(0),
		PriLim: c.PriLim,
		TxQ:    tx.NewTxHeap(),
		Ct:     atomic.NewUint32(0),
		Cap:    c.TxPCap,
	}
}

// PriMet (PriorityMet) checks to see
// if the transaction pool has enough
// cumulative priority to start mining.
func (tp *TxPool) PriMet() bool {
	return tp.CurPri.Load() >= tp.PriLim
}

// CalcPri (CalculatePriority) calculates the
// priority of a transaction by dividing the
// fees (inputs - outputs) by the size of the
// transaction and multiplying by a factor of 100.
// fees * factor / sz
func CalcPri(t *tx.Transaction) uint32 {
	if t == nil {
		fmt.Println("ERROR {TxPool.CalcPri}: received a nil transaction")
		return 0
	}
	input := t.SumInputs()
	output := t.SumOutputs()
	fees := (input - output)
	priority := (fees * 100) / t.Sz()

	if priority == 0 {
		return 1
	}
	return priority
}

// Add adds a transaction to the transaction pool.
// If the transaction pool is full, the transaction
// will not be added. Otherwise, the cumulative
// priority level is updated, the counter is
// incremented, and the transaction is added to the
// heap.
func (tp *TxPool) Add(t *tx.Transaction) {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	if t == nil {
		fmt.Println("ERROR {TxPool.Add}: received a nil transaction")
		return
	}

	if tp.Length() >= tp.Cap {
		return
	}

	tp.TxQ.Add(CalcPri(t), t)
	tp.CurPri.Add(CalcPri(t))
	tp.Ct.Add(1)
}

// ChkTxs (CheckTransactions) checks for any duplicate
// transactions in the heap and removes them.
func (tp *TxPool) ChkTxs(remover []*tx.Transaction) {
	tp.mutex.Lock()
	var count uint32
	var priority uint32

	if remover == nil {
		fmt.Printf("ERROR {TxPool.ChkTxs}: received a nil transaction")
		return
	}

	removed := tp.TxQ.Rmv(remover)

	for i := range removed {
		count++
		priority += CalcPri(removed[i])
	}

	tp.CurPri.Sub(priority)
	tp.Ct.Sub(count)

	tp.mutex.Unlock()
	return
}
