package pkg

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
)

// ChkBlk (CheckBlock) validates a block based on multiple
// conditions.
// To be valid:
// The block must be syntactically (ChkBlkSyn), semantically
// (ChkBlkSem), and configurally (ChkBlkConf) valid.
// Each transaction on the block must be syntactically (ChkTxSyn),
// semantically (ChkTxSem), and configurally (ChkTxConf) valid.
// Each transaction on the block must reference UTXO on the same
// chain (main or forked chain) and not be a double spend on that
// chain.
// Inputs:
// b *block.Block the block to be checked for validity
// Returns:
// bool True if the block is valid. false
// otherwise
func (n *Node) ChkBlk(b *block.Block) bool {
	if b == nil {
		return false
	} else if len(b.Transactions) <= 0 {
		return false
	}

	for i := range b.Transactions {
		if i == 0 && (!b.Transactions[i].IsCoinbase() || len(b.Transactions[i].Outputs) <= 0 || b.Transactions[i].SumOutputs() <= 0) {
			return false
		}
		if i != 0 && b.Transactions[i].IsCoinbase() {
			return false
		}
	}

	if !n.Chain.ChkChainsUTXO(b.Transactions[1:], b.Hdr.PrvBlkHsh) {
		return false
	}

	if b.Sz() > n.Conf.MxBlkSz {
		return false
	}

	if !b.SatisfiesPOW(b.Hdr.DiffTarg) {
		return false
	}

	return true
}

// ChkTx (CheckTransaction) validates a transaction.
// Inputs:
// t *tx.Transaction the transaction to be checked for validity
// Returns:
// bool True if the transaction is syntactically valid. false
// otherwise
func (n *Node) ChkTx(t *tx.Transaction) bool {
	for i := range t.Inputs {
		if n.Chain.IsInvalidInput(t.Inputs[i]) {
			return false
		}

		UTXO := n.Chain.GetUTXO(t.Inputs[i])

		if UTXO == nil {
			return false
		}

		if !UTXO.IsUnlckd(t.Inputs[i].UnlockingScript) {
			return false
		}
	}

	return t.Inputs != nil &&
		t.Outputs != nil &&
		len(t.Inputs) > 0 &&
		len(t.Outputs) > 0 &&
		t.SumOutputs() > 0 &&
		t.SumInputs() > 0 &&
		t.SumInputs() >= t.SumOutputs() &&
		t.Sz() <= n.Conf.MxBlkSz
}
