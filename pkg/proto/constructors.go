package proto

import (
	"unsafe"
)

// SzOfTx (SizeOfTransaction) returns the
// size of the transaction in bytes.
func SzOfTx(t *Transaction) uint32 {
	var sz uint32
	for _, txi := range t.Inputs {
		sz += uint32(unsafe.Sizeof(txi.Amount))
		sz += uint32(unsafe.Sizeof(txi.TransactionHash))
		sz += uint32(unsafe.Sizeof(txi.UnlockingScript))
		sz += uint32(unsafe.Sizeof(txi.OutputIndex))
	}
	for _, txo := range t.Outputs {
		sz += uint32(unsafe.Sizeof(txo.Amount))
		sz += uint32(unsafe.Sizeof(txo.LockingScript))
	}
	sz += uint32(unsafe.Sizeof(t.LockTime))
	sz += uint32(unsafe.Sizeof(t.Version))
	return sz
}

// NewTx (NewTransaction) returns a new
// protobuf transaction.
func NewTx(ver uint32, inpts []*TransactionInput,
	outpts []*TransactionOutput, lckTime uint32) *Transaction {
	return &Transaction{
		Version:  ver,
		Inputs:   inpts,
		Outputs:  outpts,
		LockTime: lckTime,
	}
}

// SzOfBlk (SizeOfBlock) returns the
// size of the block in bytes
// Returns:
// uint32 the number of bytes the object
// takes up
func SzOfBlk(b *Block) uint32 {
	var sz uint32
	for _, t := range b.Transactions {
		sz += SzOfTx(t)
	}
	sz += SzOfHdr(b.Header)
	return sz
}

// SzOfHdr (SizeOfBlockHeader) returns
// the size of the block header in bytes
// Returns:
// uint32 the number of bytes in the block
// header
func SzOfHdr(h *BlockHeader) uint32 {
	var sz uint32
	sz += uint32(unsafe.Sizeof(h.PrevBlockHash))
	sz += uint32(unsafe.Sizeof(h.Version))
	sz += uint32(unsafe.Sizeof(h.DifficultyTarget))
	sz += uint32(unsafe.Sizeof(h.Nonce))
	sz += uint32(unsafe.Sizeof(h.Timestamp))
	return sz
}

// NewTxInpt (NewTransactionInput) returns
// a new protobuf transaction input.

func NewTxInpt(h string, i uint32, unlckScr string, amt uint32) *TransactionInput {
	return &TransactionInput{
		TransactionHash: h,
		OutputIndex:     i,
		UnlockingScript: unlckScr,
		Amount:          amt,
	}
}

// NewTxOutpt (NewTransactionOutput) returns
// a new protobuf transaction output.
func NewTxOutpt(amt uint32, toPK string) *TransactionOutput {
	return &TransactionOutput{
		Amount:        amt,
		LockingScript: toPK,
	}
}
