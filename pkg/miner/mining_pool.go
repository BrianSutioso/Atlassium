package miner

import "BrunoCoin/pkg/block/tx"

// MiningPool is the list of transactions
// that the miner is currently mining.
type MiningPool []*tx.Transaction

// NewMiningPool selects the highest priority
// transactions from the transaction pool.
func (m *Miner) NewMiningPool() MiningPool {
	var txs []*tx.Transaction
	var blkSz uint32 = 100 // assume coinbase
	var rankings = *m.TxP.TxQ
	for i := 0; i < len(rankings); i++ {
		blkSz += rankings[i].T.Sz()
		if blkSz < m.Conf.BlkSz {
			txs = append(txs, rankings[i].T)
		} else {
			break
		}
	}
	return txs
}
