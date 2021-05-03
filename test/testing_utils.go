package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/id"
	"BrunoCoin/pkg/utils"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/phayes/freeport"
)

func GetFreePort() int {
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	return port
}

func GenConf(port int) *pkg.Config {
	c := pkg.DefaultConfig(port)
	c.CstmID = true
	c.CstmIDObj, _ = id.LoadInSmplID(blockchain.GENPK, blockchain.GENPVK)
	return c
}

// First node is always the genesis node
func NewCluster(n int) []*pkg.Node {
	cluster := []*pkg.Node{NewGenNd()}
	for i := 1; i < n; i++ {
		cluster = append(cluster, pkg.New(pkg.DefaultConfig(GetFreePort())))
	}
	return cluster
}

func StartCluster(c []*pkg.Node) {
	for _, node := range c {
		node.Start()
	}
}

func ConnectCluster(c []*pkg.Node) {
	for i := 0; i < len(c); i++ {
		for j := 0; j < len(c); j++ {
			if i == j {
				continue
			}
			c[i].ConnectToPeer(c[j].Addr)
		}
	}
}

func CheckChainLengths(t *testing.T, c []*pkg.Node, lens []int) {
	t.Helper()
	for i := 0; i < len(c); i++ {
		if c[i].Chain.Length() != lens[i] {
			t.Fatalf("%v had incorrect chain length: %v vs %v\n%v", c[i].Addr, c[i].Chain.Length(), lens[i], c[i].Chain)
		}
	}
}

func NewGenNd() *pkg.Node {
	return pkg.New(GenConf(GetFreePort()))
}

func ChkNdPrs(t *testing.T, n *pkg.Node, prs []*pkg.Node) {
	for _, pr := range prs {
		if !n.PeerDb.In(pr.Addr) {
			t.Errorf("Node didn't contain peer")
		}
	}
}

func ChkTxSeenLen(t *testing.T, n *pkg.Node, ln int) {
	if len(n.TxMap) != ln {
		t.Errorf("Failed: Node was expected to see %v txs, but has only seen %v", ln, len(n.TxMap))
	}
}

func ChkMnChnLen(t *testing.T, n *pkg.Node, l int) {
	t.Helper()
	if n.Chain.Length() != l {
		t.Errorf("Failed: Node was expected to have a main chain of length %v, but had one of length %v\n", l, n.Chain.Length())
	}
}

func ChkMnChnCons(t *testing.T, ns []*pkg.Node) {
	t.Helper()
	n1Chn := ns[0].Chain
	for i := 1; i < len(ns); i++ {
		ChkEqBlks(t, n1Chn.List(), ns[i].Chain.List())
	}
}

func ChkEqBlks(t *testing.T, c1 []*block.Block, c2 []*block.Block) {
	t.Helper()
	if len(c1) != len(c2) {
		t.Errorf("Failed: One node had a main chain length of %v, while another node had %v\n", len(c1), len(c2))
	}
	for i := 0; i < len(c1); i++ {
		if c1[i].Hash() != c2[i].Hash() {
			t.Errorf("Failed: At idx %v, one node's block hash was %v, while another node's block hash was %v", i, c1[i].Hash(), c2[i].Hash())
		}
	}
}

func NwDmyRPCSrv() *pkg.Node {
	n := pkg.New(pkg.NilConfig(GetFreePort()))
	return n
}
func StrtDmyRPCSrv(n *pkg.Node) {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addr := fmt.Sprintf("%v:%v", hostname, n.Conf.Port)
	n.Addr = addr
	n.PeerDb.SetAddr(addr)
	utils.Debug.Printf("%v {MALICIOUS} started\n", utils.FmtAddr(addr))
	n.StartServer(addr)
}

func SndDmyTx(n *pkg.Node, t *tx.Transaction) {
	n.TxMap[t.Hash()] = true
	for _, a := range n.PeerDb.List() {
		utils.Debug.Printf("%v {MALICIOUS} sending %v to peer %v.\n", utils.FmtAddr(n.Addr), t.NameTag(), utils.FmtAddr(a.Addr.Addr))
		a.Addr.ForwardTransactionRPC(t.Serialize())
	}
}

func AsrtBal(t *testing.T, n *pkg.Node, a uint32) {
	pk := hex.EncodeToString(n.Id.GetPublicKeyBytes())
	if n.GetBalance(pk) != a {
		t.Errorf("Failed: Node {%v} was expected to have a balance of %v, but had a balance of %v\n", n.Addr, a, n.GetBalance(pk))
	}
}
