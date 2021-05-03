package tx

import (
	"BrunoCoin/pkg/block/tx/txi"
	"BrunoCoin/pkg/block/tx/txo"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"fmt"
	"strconv"
	"strings"
)

// Transaction is a wrapper around the
// protobuf transaction so methods
// can be called on it and additional fields
// may be added.
// Inputs represent the same transaction inputs
// in the protobuf, but these ones are wrapped.
// Outputs represent the same transaction outputs
// in the protobuf, but these ones are wrapped.
type Transaction struct {
	Version  uint32
	Inputs   []*txi.TransactionInput
	Outputs  []*txo.TransactionOutput
	LockTime uint32
}

// Sz (Size) returns the size of the
// underlying protobuf transaction.
// Returns:
// uint32	size in bytes of the underlying
// protobuf transaction.
func (t *Transaction) Sz() uint32 {
	return proto.SzOfTx(t.Serialize())
}

// SumInputs returns the sum of the inputs.
// Returns:
// uint32	the sum of the amounts on each
// input
func (t *Transaction) SumInputs() uint32 {
	var r uint32 = 0
	for _, v := range t.Inputs {
		r += v.Amount
	}
	return r
}

// SumOutputs returns the sum of the outputs.
// Returns:
// uint32	the sum of the amounts on each
// output
func (t *Transaction) SumOutputs() uint32 {
	var r uint32 = 0
	for _, v := range t.Outputs {
		r += v.Amount
	}
	return r
}

// Hash returns the hash of the underlying
// protobuf transaction.
// Returns:
// string	the hash of the transaction
// represented as a hex string
func (t *Transaction) Hash() string {
	pureInputs := make([]string, 0)
	for _, i := range t.Inputs {
		pureInputs = append(pureInputs, i.Hash())
	}
	pureOutputs := make([]string, 0)
	for _, o := range t.Outputs {
		pureOutputs = append(pureOutputs, o.Hash())
	}
	pureData := []byte(fmt.Sprintf("%v/%v/%v/%v", t.Version, t.LockTime, strings.Join(pureInputs, "/"), strings.Join(pureOutputs, "/")))
	return utils.Hash(pureData)
}

// IsCoinbase returns whether or not the
// transaction is a coinbase transaction.
// Returns:
// bool	true if the transaction is a coinbase.
// False otherwise.
func (t *Transaction) IsCoinbase() bool {
	return len(t.Inputs) == 0
}

// Deserialize creates an identical transaction
// from a passed in protobuf transaction.
// Inputs:
// ptx	*proto.Transaction a pointer to a protobuf
// transaction
// Returns:
// *Transaction a pointer to a transaction that has
// identical fields (ish) as the protobuf transaction
func Deserialize(ptx *proto.Transaction) *Transaction {
	inputs := make([]*txi.TransactionInput, len(ptx.Inputs))
	for i := range inputs {
		inputs[i] = &txi.TransactionInput{
			TransactionHash: ptx.Inputs[i].TransactionHash,
			OutputIndex:     ptx.Inputs[i].OutputIndex,
			UnlockingScript: ptx.Inputs[i].UnlockingScript,
			Amount:          ptx.Inputs[i].Amount,
		}
	}
	outputs := make([]*txo.TransactionOutput, len(ptx.Outputs))
	for i := range outputs {
		outputs[i] = &txo.TransactionOutput{
			Amount:        ptx.Outputs[i].Amount,
			LockingScript: ptx.Outputs[i].LockingScript,
		}
	}
	return &Transaction{Version: ptx.Version, Inputs: inputs,
		Outputs: outputs, LockTime: ptx.LockTime}
}

// Serialize serializes the transaction into a
// protobuf transaction so that it can be broadcast
// to the network.
// Returns:
// *proto.Transaction	pointer to a protobuf
// transaction with identical fields as the transaction
func (t *Transaction) Serialize() *proto.Transaction {
	ins := make([]*proto.TransactionInput, len(t.Inputs))
	for i := range ins {
		ins[i] = t.Inputs[i].Serialize()
	}
	outs := make([]*proto.TransactionOutput, len(t.Outputs))
	for i := range outs {
		outs[i] = t.Outputs[i].Serialize()
	}
	return &proto.Transaction{
		Version:  t.Version,
		LockTime: t.LockTime,
		Inputs:   ins,
		Outputs:  outs,
	}
}

func (t *Transaction) NameTag() string {
	i, _ := strconv.ParseInt(t.Hash()[:10], 16, 64)
	return fmt.Sprintf("%v", utils.Colorize(fmt.Sprintf("tx-%v", t.Hash()[:6]), int(i)))
}
