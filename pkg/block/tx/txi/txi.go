package txi

import (
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"fmt"
)

// TransactionInput is a wrapper around the
// protobuf transaction input so methods
// can be called on it and additional fields
// may be added.
type TransactionInput struct {
	TransactionHash string
	OutputIndex     uint32
	UnlockingScript string
	Amount          uint32
}

// Serialize serializes a transaction input to a protobuf
// version of a transaction input that can be broadcast
// on the network.
// Returns:
// *proto.TransactionInput	pointer to a protobuf
// transaction input.
func (txi *TransactionInput) Serialize() *proto.TransactionInput {
	return &proto.TransactionInput{
		TransactionHash: txi.TransactionHash,
		OutputIndex:     txi.OutputIndex,
		UnlockingScript: txi.UnlockingScript,
		Amount:          txi.Amount,
	}
}

// Deserialize serializes a protobuf transaction to a normal
// transaction so it can be properly tested
// Returns:
// *	pointer to a
// transaction input.
func Deserialize(inp *proto.TransactionInput) *TransactionInput {
	ip := &TransactionInput{
		TransactionHash: inp.TransactionHash,
		UnlockingScript: inp.UnlockingScript,
		OutputIndex:     inp.OutputIndex,
		Amount:          inp.Amount,
	}
	return ip
}

func (txi *TransactionInput) Hash() string {
	pureData := []byte(fmt.Sprintf("%v/%v/%v/%v", txi.TransactionHash, txi.OutputIndex, txi.UnlockingScript, txi.Amount))
	return utils.Hash(pureData)
}
