package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/utils"
	"testing"
	"time"
)

func TestNoUTXO(t *testing.T){
	utils.SetDebug(true)
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Start()
	node2.Start()
	genNd.StartMiner()
	node2.StartMiner()
	genNd.ConnectToPeer(node2.Addr)
	if peer := genNd.PeerDb.Get(node2.Addr); peer == nil {
		t.Fatal("Seed node did not contain newNode as peer")
	}
	if peer := node2.PeerDb.Get(genNd.Addr); peer == nil {
		t.Fatal("New node did not contain seedNode as peer")
	}
	// Sleep to give time for both nodes to connect
	time.Sleep(1 * time.Second)

	// Checks that both nodes have the same main blockchain,
	// and are both connected
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
	ChkNdPrs(t, genNd, []*pkg.Node{node2})
	ChkNdPrs(t, node2, []*pkg.Node{genNd})

	// The genesis node sends money to node 2 with a high
	// fee
	node2.SendTx(100, 100, genNd.Id.GetPublicKeyBytes())

	// Sleep to give time for the transaction to be broadcast,
	// and validated by the other node
	time.Sleep(6 * time.Second)

	/* node2.SendTx(1, 50, genNd.Id.GetPublicKeyBytes())
	node2.SendTx(3, 30, genNd.Id.GetPublicKeyBytes())
	node2.SendTx(4, 20, genNd.Id.GetPublicKeyBytes()) */
	// Check that both nodes have "seen" 1 transaction
	ChkTxSeenLen(t, genNd, 0)
	ChkTxSeenLen(t, node2, 0)
	time.Sleep(6 * time.Second)

	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
}

func TestSendLiminal(t *testing.T){
	utils.SetDebug(true)
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Start()
	node2.Start()
	genNd.ConnectToPeer(node2.Addr)
	if peer := genNd.PeerDb.Get(node2.Addr); peer == nil {
		t.Fatal("Seed node did not contain newNode as peer")
	}
	if peer := node2.PeerDb.Get(genNd.Addr); peer == nil {
		t.Fatal("New node did not contain seedNode as peer")
	}
	// Sleep to give time for both nodes to connect
	time.Sleep(1 * time.Second)

	// Checks that both nodes have the same main blockchain,
	// and are both connected
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
	ChkNdPrs(t, genNd, []*pkg.Node{node2})
	ChkNdPrs(t, node2, []*pkg.Node{genNd})

	// The genesis node sends money to node 2 with a high
	// fee
	genNd.SendTx(100, 100, node2.Id.GetPublicKeyBytes())
	node2.SendTx(100, 100, genNd.Id.GetPublicKeyBytes())
	// Sleep to give time for the transaction to be broadcast,
	// and validated by the other node
	time.Sleep(6 * time.Second)
	node2.StartMiner()

	/* node2.SendTx(1, 50, genNd.Id.GetPublicKeyBytes())
	node2.SendTx(3, 30, genNd.Id.GetPublicKeyBytes())
	node2.SendTx(4, 20, genNd.Id.GetPublicKeyBytes()) */
	// Check that both nodes have "seen" 1 transaction
	ChkTxSeenLen(t, genNd, 1)
	ChkTxSeenLen(t, node2, 1)
	time.Sleep(6 * time.Second)

	// making sure the chains are the same
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
}