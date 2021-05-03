package wallet

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/id"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"encoding/hex"
	"fmt"
	"sync"
)

// TxReq (TransactionRequest) that represents
// the minimum amount of information needed
// to make a transaction.
// PubK (PublicKey) represents the serialized
// public key of the person they want to pay.
// Amt (Amount) represents the amount of money
// they want to pay the person.
type TxReq struct {
	PubK []byte
	Amt  uint32
	Fee  uint32
}

// Wallet provides the functionality to make
// transactions from transaction requests and
// send them to the node to be broadcast on
// the network.
// Conf represents the configuration for the
// wallet.
// Id represents the identity of the person
// using the wallet.
// Chain represents the blockchain, as the
// wallet needs to be able to query the chain
// for enough UTXO to fulfill a transaction request.
// SendTx (SendTransaction) is a channel for sending
// fulfilled transaction requests (now in the form of
// a transaction) to the node, in order to be sent
// across the network.
// LmnlTxs (LiminalTransactions) represent the
// transactions that the wallet has made, but that
// do not have enough proof of work on top of them
// to be considered valid by everyone.
// Mut (Mutex) is a mutex for concurrent accesses
// to non-atomic reads/writes for the struct
type Wallet struct {
	Conf    *Config
	Id      id.ID
	Chain   *blockchain.Blockchain
	SendTx  chan *tx.Transaction
	LmnlTxs *LiminalTxs
	Addr    string

	mutex sync.Mutex
}

// SetAddr (SetAddress) sets the address
// of the node in the wallet.
func (w *Wallet) SetAddr(a string) {
	w.mutex.Lock()
	w.Addr = a
	w.mutex.Unlock()
}

// New creates a wallet object.
// Inputs:
// c *Config the configuration
// for the wallet
// id id.ID the id of the node
// chain *blockchain.Blockchain the
// blockchain that the wallet needs a
// references to find UTXO for making transactions.
// Returns:
// *Wallet the new wallet object
func New(c *Config, id id.ID, chain *blockchain.Blockchain) *Wallet {
	if !c.HasWt {
		return nil
	}
	return &Wallet{
		Conf:    c,
		Id:      id,
		Chain:   chain,
		SendTx:  make(chan *tx.Transaction),
		LmnlTxs: NewLmnlTxs(c),
	}
}

// HndlBlk (HandleBlock) is called after a new
// block is added to the main chain. However, the
// inputted block is a "safe block amount" down from
// the top block of the main chain. If the wallet
// has still not seen some of its transactions added
// to the main chain this far down, then it may have
// to resend the transactions out.
// Inputs:
// b *block.Block the block that is "safe block amount"
// down from the last block on the main chain
func (w *Wallet) HndlBlk(b *block.Block) {
	if len(b.Transactions) == 0 || b == nil {
		return
	}

	abvThreshold, _ := w.LmnlTxs.ChkTxs(b.Transactions)

	if len(abvThreshold) <= 0 || abvThreshold == nil {
		return
	}

	for i := range abvThreshold {
		if abvThreshold[i] == nil {
			continue
		}

		abvThreshold[i].LockTime += 1
		w.SendTx <- abvThreshold[i]
		w.LmnlTxs.Add(abvThreshold[i])

		utils.Debug.Printf("Address " + utils.FmtAddr(w.Addr) + " -> transaction " + abvThreshold[i].NameTag())
	}

	return
}

// HndlTxReq (HandleTransactionRequest) attempts to
// create a transaction from the request, as well as
// sending this transaction to the node to be forwarded
// on the network. It generates the transaction by first
// asking the blockchain for enough UTXO to construct the
// transaction. At this point, the transaction is made, but
// not valid by the consensus since it is not mined onto the
// main chain and have enough POW on top of it. Therefore,
// we must add it to our liminal transactions (transactions that
// have been made/broadcast but not validated).
// Inputs:
// txR *TxReq a transaction request from the node
func (w *Wallet) HndlTxReq(txR *TxReq) {
	if txR.Amt == 0 {
		return
	}

	var protoTxI []*proto.TransactionInput
	var protoTxO []*proto.TransactionOutput

	UTXOinfo, change, enough := w.Chain.GetUTXOForAmt(txR.Amt+txR.Fee, hex.EncodeToString(w.Id.GetPublicKeyBytes()))

	if !enough {
		return
	}

	for i := range UTXOinfo {
		sig, Error := UTXOinfo[i].UTXO.MkSig(w.Id)
		if Error != nil {
			fmt.Println("ERROR {m.HndlTxReq}: unable to generate locking script")
			return
		}
		protoTxI = append(protoTxI, proto.NewTxInpt(UTXOinfo[i].TxHsh, UTXOinfo[i].OutIdx, sig, UTXOinfo[i].Amt))
	}

	protoTxO = append(protoTxO, proto.NewTxOutpt(txR.Amt, hex.EncodeToString(txR.PubK)))
	if change > 0 {
		protoTxO = append(protoTxO, proto.NewTxOutpt(change, hex.EncodeToString(w.Id.GetPublicKeyBytes())))
	}

	protoTx := proto.NewTx(w.Conf.TxVer, protoTxI, protoTxO, w.Conf.DefLckTm)
	Tx := tx.Deserialize(protoTx)

	w.LmnlTxs.Add(Tx)
	w.SendTx <- Tx

	utils.Debug.Printf("Address " + utils.FmtAddr(w.Addr) + " -> transaction " + Tx.NameTag())

	return
}
