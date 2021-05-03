package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/utils"
	"testing"
	"time"
)

// TestBootstrapEmpty starts two nodes. It
// immediately bootstraps node 2 when there is
// nothing to bootstrap.
func TestBootstrapEmpty(t *testing.T) {
	utils.SetDebug(true)

	// This creates the genesis node as well as
	// creating one other node in the network.
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Start()
	node2.Start()

	// Node connects to the network
	genNd.ConnectToPeer(node2.Addr)

	// Sleep to give time for both nodes to connect
	time.Sleep(1 * time.Second)

	// Bootstrap Node 2. There should be nothing to
	// bootstrap
	err := node2.Bootstrap()
	if err != nil {
		t.Errorf("Test errored when attempting to" +
			"bootstrap.\n")
	}

	// Sleep to give enough time for the bootstrapping
	// requests and responses to take place.
	time.Sleep(time.Second * 2)

	// Checks to see that both miners main chains are
	// the same and that they only have the genesis
	// block
	ChkMnChnLen(t, genNd, 1)
	ChkMnChnLen(t, genNd, 1)
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
}

// TestBootstrapTwoNodesTwoBlocks creates two nodes. It only
// starts one of them and starts the miner for that node (genNd).
// Then, a transaction is made, a block is mined and not broadcast
// to the other node. Then, the other node is started and
// bootstrapped. Then tested to make sure they are in consensus.
func TestBootstrapTwoNodesTwoBlocks(t *testing.T) {
	utils.SetDebug(true)

	// Creates two nodes, but only starts the genesis node.
	// Also starts the miner for the genesis node. Sets the
	// mining difficulty target for the genesis node very
	// easy so that blocks can be mined quickly
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)
	genNd.Start()
	genNd.StartMiner()

	// Transaction made from genesis node with large fee, so
	// that it is immediately mined
	genNd.SendTx(50, 100, node2.Id.GetPublicKeyBytes())

	// Sleep to give time for the transaction to be mined
	// to a block
	time.Sleep(time.Second * 3)

	// Checks to see that the transaction was mined to a block
	// and that the other node's length is still only 1
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 1)

	// Starts the other node and connects it to the network
	node2.Start()
	genNd.ConnectToPeer(node2.Addr)

	// Sleep to give time for the node to connect to the network
	time.Sleep(time.Second * 1)

	// Bootstrap the node so that it can catch up to the genesis
	// node
	err := node2.Bootstrap()
	if err != nil {
		t.Errorf("Test errored when attempting to" +
			"bootstrap.\n")
	}

	// Sleep to give enough time for the bootstrapping process
	time.Sleep(time.Second * 2)

	// Check to see that both nodes have the same length
	// main chain now
	ChkMnChnLen(t, genNd, 2)
	ChkMnChnLen(t, node2, 2)
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
}

// TestBootstrapTwoNodesManyBlocks makes two nodes but only
// starts the genesis node. It then creates 4 transactions
// and mines each one to a block. Then the second node
// starts up and bootstraps.
func TestBootstrapTwoNodesManyBlocks(t *testing.T) {
	utils.SetDebug(true)

	// Two nodes created. Only genesis node started and its
	// miner is started as well with an easy difficulty for
	// fast mining.
	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)
	genNd.Start()
	genNd.StartMiner()

	// Four transactions are made and mined to the blockchain
	for i := 0; i < 4; i++ {
		genNd.SendTx(10, 50, node2.Id.GetPublicKeyBytes())
		// Sleep to give time for the transaction to be mined
		time.Sleep(time.Second * 3)
	}

	// Check to see that the genesis node has a chain of length
	// 5 while node2 has length of 1
	ChkMnChnLen(t, genNd, 5)
	ChkMnChnLen(t, node2, 1)

	// Node 2 finally starts and connects to the network
	node2.Start()
	genNd.ConnectToPeer(node2.Addr)

	// Sleep to give time for the node to connect
	time.Sleep(time.Second * 1)

	// Bootstrap the second node, it should receive 4 blocks
	err := node2.Bootstrap()
	if err != nil {
		t.Errorf("Test errored when attempting to" +
			"bootstrap.\n")
	}

	// Sleep to give time for node2 to bootstrap
	time.Sleep(time.Second * 1)

	// Check to see that the miners are in consensus.
	ChkMnChnLen(t, genNd, 5)
	ChkMnChnLen(t, node2, 5)
	ChkMnChnCons(t, []*pkg.Node{genNd, node2})
}

func TestBootstrapDeadPeer(t *testing.T) {
	utils.SetDebug(true)
	nodes := NewCluster(3)
	StartCluster(nodes)
	ConnectCluster(nodes)
	// Node 2 will bootstrap from the genesis block
	nodes[2].PauseNetwork()
	nodes[0].StartMiner()
	nodes[0].SendTx(800, 50, nodes[1].Id.GetPublicKeyBytes())
	time.Sleep(time.Second * 5)
	// Node 1 will be paused
	nodes[1].PauseNetwork()
	nodes[0].SendTx(200, 50, nodes[2].Id.GetPublicKeyBytes())
	time.Sleep(time.Second * 5)
	CheckChainLengths(t, nodes, []int{3, 2, 1})
	nodes[2].ResumeNetwork()
	if nodes[2].Bootstrap() != nil {
		t.Errorf("Test errored when attempting to" +
			"bootstrap.\n")
	}
	time.Sleep(time.Second * 5)
	CheckChainLengths(t, nodes, []int{3, 2, 3})
}

func TestBootstrapCompetingChains(t *testing.T) {
	utils.SetDebug(true)
	nodes := NewCluster(3)
	StartCluster(nodes)
	ConnectCluster(nodes)
	// Node 2 will bootstrap from the genesis block
	nodes[2].PauseNetwork()
	nodes[0].StartMiner()
	nodes[0].SendTx(800, 50, nodes[1].Id.GetPublicKeyBytes())
	time.Sleep(time.Second * 5)
	nodes[1].SendTx(200, 50, nodes[2].Id.GetPublicKeyBytes())
	time.Sleep(time.Second * 5)
	// Node 1 will be behind by a block
	nodes[1].PauseNetwork()
	nodes[0].SendTx(200, 50, nodes[2].Id.GetPublicKeyBytes())
	time.Sleep(time.Second * 5)
	CheckChainLengths(t, nodes, []int{4, 3, 1})
	nodes[1].ResumeNetwork()
	nodes[2].ResumeNetwork()
	if nodes[2].Bootstrap() != nil {
		t.Errorf("Test errored when attempting to" +
			"bootstrap.\n")
	}
	time.Sleep(time.Second * 5)
	CheckChainLengths(t, nodes, []int{4, 3, 4})
}
