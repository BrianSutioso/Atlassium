package block

import (
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// Header is a wrapper around the
// protobuf block header for convenience
// of adding extra fields and calling
// methods.
// Ver the software version of the node
// PrvBlkHsh the hash of the previous
// block that the block the header belongs
// to is after. Represented as a hex string.
// MrklRt is the merkle root of all the
// transactions within the block. Represented
// as a hex string.
// Timestamp is the timestamp of when the
// block was made represented in seconds
// of UNIX time.
// DiffTarg is the difficulty target that
// needs to be met for the block to be mined.
// Nonce is the nonce that was found for the
// block that meets the difficulty target.
type Header struct {
	Ver       uint32
	PrvBlkHsh string
	MrklRt    string
	Timestamp uint32
	DiffTarg  string
	Nonce     uint32
}

// Block is a wrapper around the
// protobuf block for convenience
// of adding extra fields and calling
// methods.
// Hdr is the block header
// Transactions is all transactions
// on the block.
type Block struct {
	Hdr          Header
	Transactions []*tx.Transaction
}

func New(prvHsh string, txs []*tx.Transaction, target string) *Block {
	txsd := make([]*proto.Transaction, len(txs))
	for i := range txsd {
		txsd[i] = txs[i].Serialize()
	}
	b := &proto.Block{
		Header: &proto.BlockHeader{
			Version:          0,
			PrevBlockHash:    prvHsh,
			MerkleRoot:       CalcMrklRt(txs),
			DifficultyTarget: target,
		},
		Transactions: txsd,
	}
	return Deserialize(b)
}

// Deserialize creates a new Block by taking in
// a protobuf block and wrapping it.
func Deserialize(b *proto.Block) *Block {
	txs := make([]*tx.Transaction, len(b.Transactions))
	for i := range txs {
		txs[i] = tx.Deserialize(b.Transactions[i])
	}
	return &Block{
		Hdr: Header{
			Ver:       b.Header.Version,
			PrvBlkHsh: b.Header.PrevBlockHash,
			MrklRt:    b.Header.MerkleRoot,
			Timestamp: b.Header.Timestamp,
			DiffTarg:  b.Header.DifficultyTarget,
			Nonce:     b.Header.Nonce,
		},
		Transactions: txs,
	}
}

// Serialize unwraps the block and returns
// the underlying protobuf block.
func (b *Block) Serialize() *proto.Block {
	txs := make([]*proto.Transaction, len(b.Transactions))
	for i := range txs {
		txs[i] = b.Transactions[i].Serialize()
	}
	return &proto.Block{
		Header: &proto.BlockHeader{
			Version:          b.Hdr.Ver,
			PrevBlockHash:    b.Hdr.PrvBlkHsh,
			MerkleRoot:       b.Hdr.MrklRt,
			Timestamp:        b.Hdr.Timestamp,
			DifficultyTarget: b.Hdr.DiffTarg,
			Nonce:            b.Hdr.Nonce,
		},
		Transactions: txs,
	}
}

// SatisfiesPOW tests a hash and a difficulty
// target to see if the hash satisfies the
// difficulty target (hash < dif target).
// Inputs:
// dt	string	represents the difficulty
// target as a hex string
// Returns:
// bool True if the hash was less than the
// difficulty target, false otherwise.
func (b *Block) SatisfiesPOW(dt string) bool {
	hsh, err := hex.DecodeString(b.Hash())
	if err != nil {
		fmt.Printf("ERROR {Block.SatisfiesPOW}: "+
			"Could not decode block {%v}"+
			"-> error in Block.Hash().\n", b.Hash())
		return false
	}
	difTrg, err := hex.DecodeString(dt)
	if err != nil {
		fmt.Printf("ERROR {Block.SatisfiesPOW}: "+
			"Could not decode difficulty target {%v}.\n", dt)
		return false
	}
	return bytes.Compare(hsh, difTrg) == -1
}

// Sz (Size) returns the size of the
// block in bytes
// Returns:
// uint32 the number of bytes the block
// takes up
func (b *Block) Sz() uint32 {
	return proto.SzOfBlk(b.Serialize())
}

// CalcMrklRt (CalculateMerkleRoot) calculates
// the merkle root for a list of transactions.
// Look up merkle trees for further description.
// Input:
// txs	[]*tx.Transaction a list of transactions
// that represent the leaves of the merkle tree.
// Returns:
// string	the root of the merkle tree represented
// as a hex string.
func CalcMrklRt(txs []*tx.Transaction) string {
	var hshs []string
	if len(txs) > 1 && len(txs)%2 != 0 {
		txs = append(txs, txs[len(txs)-1])
	}
	for _, t := range txs {
		hshs = append(hshs, t.Hash())
	}
	for len(hshs) != 1 {
		var newHshs []string
		if len(hshs)%2 != 0 {
			hshs = append(hshs, hshs[len(hshs)-1])
		}
		for i := 0; i < len(hshs); i += 2 {
			byts1, _ := hex.DecodeString(hshs[i])
			byts2, _ := hex.DecodeString(hshs[i+1])
			byts3 := append(byts1[:], byts2[:]...)
			newHsh := utils.Hash(byts3)
			newHshs = append(newHshs, newHsh)
		}
		hshs = newHshs
	}
	if hshs == nil || len(hshs) < 1 {
		fmt.Printf("ERROR {block.CaclMrkRt}: function" +
			"ended up not being able to calculate a root.\n")
		return ""
	}
	return hshs[0]
}

func (b *Block) String() string {
	return fmt.Sprintf("%v", b.Hash())
}

// Hash returns the hash of a block.
// Returns:
// string	the hash of the block represented
// as a hex string
func (b *Block) Hash() string {
	pureData := []byte(fmt.Sprintf("%v", b.Hdr))
	return utils.Hash(pureData)
}

func (b *Block) NameTag() string {
	i, _ := strconv.ParseInt(b.Hash()[:10], 16, 64)
	return fmt.Sprintf("%v", utils.Colorize(fmt.Sprintf("block-%v", b.Hash()[:8]), int(i)))
}

func (b *Block) Summarize() string {
	txs := make([]string, 0)
	for _, t := range b.Transactions {
		txs = append(txs, t.NameTag())
	}
	i, _ := strconv.ParseInt(b.Hdr.PrvBlkHsh[:10], 16, 64)
	prevNameTag := utils.Colorize(fmt.Sprintf("block-%v", b.Hdr.PrvBlkHsh[:6]), int(i))
	return fmt.Sprintf("{prev: %v, txs: [%v]}", prevNameTag, strings.Join(txs, ", "))
}
