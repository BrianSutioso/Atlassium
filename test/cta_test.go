package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/miner"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"fmt"
	"testing"
	"time"
)

// Two nodes created. Genesis node makes transaction, followed
// by another transaction attempted to be made (invalid
// due to not enough utxo). Other node starts mining and
// mines the block and shares it. A malicious node enters
// the network and attempts to send an invalid transaction
// that uses an old UTXO and invalid signing process.
func TestBasicSystem(t *testing.T) {
	utils.SetDebug(true)

	// This creates the genesis node as well as
	// creating one other node in the network. The genesis
	// node is given an easy difficulty so that it mines
	// quickly since it is uncontested anyways.
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNd.Start()
	node2.Start()

	// Node connects to the network
	genNd.ConnectToPeer(node2.Addr)

	// Sleep to give time for both nodes to connect
	time.Sleep(1 * time.Second)

	// Checks that both nodes have the same main blockchain,
	// and are both connected
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
	ChkNdPrs(t, genNd, []*pkg.Node{node2})
	ChkNdPrs(t, node2, []*pkg.Node{genNd})

	// The genesis node sends money to node 2 with a high
	// fee so that a miner (once started) will immediately
	// mine it to a block.
	genNd.SendTx(10, 50, node2.Id.GetPublicKeyBytes())

	// Sleep to give time for the transaction to be broadcast,
	// and validated by the other node
	time.Sleep(1 * time.Second)

	// Check that both nodes have "seen" 1 transaction
	ChkTxSeenLen(t, genNd, 1)
	ChkTxSeenLen(t, node2, 1)

	// The following transaction should not be able to be made,
	// since the first transaction has not been mined to a block;
	// therefore, the only utxo available is currently in use
	// (but not mined)
	genNd.SendTx(10, 50, node2.Id.GetPublicKeyBytes())

	// Sleep to give time for the transaction to be made,
	// potentially (incorrectly) be broadcast, and potentially
	// (incorrectly) validated
	time.Sleep(1 * time.Second)

	// Check that both nodes have "seen" 1 transaction since
	// the second transaction should have never been made or
	// broadcast
	ChkTxSeenLen(t, genNd, 1)
	ChkTxSeenLen(t, node2, 1)

	// Node2 starts mining the one transaction it should have
	// in its pool
	node2.StartMiner()

	// Sleep to give enough time for node2 to mine the
	// transaction to a block, the block to be broadcast,
	// the block to be validated, and the block to be
	// added to each node's blockchain.
	time.Sleep(3 * time.Second)

	// Checks that each node's main chain is 2 blocks
	// long and that they are also the exact same
	// main chains on both nodes.
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 2)
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})

	// Malicious node started that just has the
	// capability of interacting with the network.
	malNd := NwDmyRPCSrv()
	StrtDmyRPCSrv(malNd)

	// Malicious node connects to the network
	malNd.ConnectToPeer(genNd.Addr)

	// Sleep to give enough time for the malicious node
	// to connect to the network
	time.Sleep(1 * time.Second)

	// Checks to see that the malicious node is connected
	// to the genesis node
	ChkNdPrs(t, genNd, []*pkg.Node{node2, malNd})
	ChkNdPrs(t, malNd, []*pkg.Node{genNd})

	// Creates an invalid transaction of the malicious node
	// trying to spend the money given to the genesis node
	// in the genesis transaction.
	tforinp := genNd.Chain.LastBlock.PrevNode.Block.Transactions[0]
	unlck, _ := tforinp.Outputs[0].MkSig(genNd.Id)
	txi := []*proto.TransactionInput{
		proto.NewTxInpt(tforinp.Hash(), 0, unlck, tforinp.Outputs[0].Amount),
	}
	amt2 := tforinp.Outputs[0].Amount - 10 - 40
	txo := []*proto.TransactionOutput{
		proto.NewTxOutpt(10, fmt.Sprintf("%x", malNd.Id.GetPublicKeyBytes())),
		proto.NewTxOutpt(amt2, fmt.Sprintf("%x", malNd.Id.GetPublicKeyBytes())),
	}
	txx := tx.Deserialize(proto.NewTx(0, txi, txo, 0))

	// Malicious node sends the invalid transaction to the
	// network. This invalid transaction will be treated as
	// an orphan.
	SndDmyTx(malNd, txx)

	// Sleep to give time for the transaction to be sent to
	// the honest nodes, classify the transaction as an orphan,
	// potentially (incorrectly) mine the transaction
	// to a block, potentially (incorrectly) broadcast/validate the
	// block and add it
	time.Sleep(3 * time.Second)

	// Checks to see that each honest node in the network
	// still has the same main chain length and has the
	// exact same main chains. Also asserts that they have
	// the correct balances.
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 2)
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})

	// The genesis node should have the amount of money
	// generated in the genesis transaction subtracted by
	// the transaction amount + fee for the first/only transaction
	// it made
	// The other node should have the transaction amount that
	// the genesis node paid to it + the fees associated with it
	// since it mined the transaction as well. Lastly it should
	// also have the minting reward amount added to it's balance.
	AsrtBal(t, genNd, blockchain.DefaultConfig().InitSbsdy-60)
	AsrtBal(t, node2, 50+10+miner.DefaultConfig(4).InitSubsdy)
}
