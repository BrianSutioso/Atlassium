package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/block/tx/txi"
	"BrunoCoin/pkg/block/tx/txo"
	"BrunoCoin/pkg/miner"
	"BrunoCoin/pkg/utils"
	"encoding/hex"
	"testing"
	"time"
)

func CreateTx(n *pkg.Node, toPK []byte, factor uint32) *tx.Transaction {
	inUTXO1 := &txo.TransactionOutput{
		Amount:        200 * factor,
		LockingScript: hex.EncodeToString(n.Id.GetPublicKeyBytes()),
	}

	sig1, _ := inUTXO1.MkSig(n.Id)

	TxI1 := &txi.TransactionInput{
		TransactionHash: inUTXO1.Hash(),
		OutputIndex:     0,
		UnlockingScript: sig1,
		Amount:          inUTXO1.Amount,
	}

	inUTXO2 := &txo.TransactionOutput{
		Amount:        400 * factor,
		LockingScript: hex.EncodeToString(n.Id.GetPublicKeyBytes()),
	}

	sig2, _ := inUTXO2.MkSig(n.Id)

	TxI2 := &txi.TransactionInput{
		TransactionHash: inUTXO2.Hash(),
		OutputIndex:     1,
		UnlockingScript: sig2,
		Amount:          inUTXO2.Amount,
	}

	TxO1 := &txo.TransactionOutput{
		Amount:        200 * factor,
		LockingScript: hex.EncodeToString(toPK),
	}

	TxO2 := &txo.TransactionOutput{
		Amount:        100 * factor,
		LockingScript: hex.EncodeToString(n.Id.GetPublicKeyBytes()),
	}

	return &tx.Transaction{
		Version:  n.Wallet.Conf.TxVer,
		Inputs:   []*txi.TransactionInput{TxI1, TxI2},
		Outputs:  []*txo.TransactionOutput{TxO1, TxO2},
		LockTime: n.Wallet.Conf.DefLckTm,
	}
}

func TestNoDupsChkTxs(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	newNode := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()
	genNode.ConnectToPeer(newNode.Addr)

	tx1 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 3)
	tx2 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 4)
	tx3 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 5)

	genNode.Mnr.TxP.Add(tx1)
	genNode.Mnr.TxP.Add(tx2)
	genNode.Mnr.TxP.Add(tx3)

	tx4 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 6)
	tx5 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 7)
	tx6 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 8)

	txArray := []*tx.Transaction{tx4, tx5, tx6}

	genNode.Mnr.TxP.ChkTxs(txArray)

	if genNode.Mnr.TxP.Ct.Load() != 3 {
		t.Errorf("Expected: %d - Actual: %d", 3, genNode.Mnr.TxP.Ct.Load())
	}
}

func TestHndlTxReqAmtZero(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	newNode := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()
	genNode.SendTx(0, 0, newNode.Id.GetPublicKeyBytes())

	ChkTxSeenLen(t, genNode, 0)
	ChkTxSeenLen(t, newNode, 0)
}

func TestChkTxsDuplicates(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	newNode := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()
	genNode.ConnectToPeer(newNode.Addr)

	tx1 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 3)
	tx2 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 3)
	tx3 := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 3)

	txArray := []*tx.Transaction{tx1, tx2, tx3}

	genNode.Mnr.TxP.ChkTxs(txArray)

	if genNode.Mnr.TxP.Ct.Load() != 0 {
		t.Errorf("Expected %d - Actual: %d", 0, genNode.Mnr.TxP.Ct.Load())
	}
}

func TestHndlNotChangedTx(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	newNode := pkg.New(pkg.DefaultConfig(GetFreePort()))

	genNode.Start()
	newNode.Start()
	genNode.ConnectToPeer(newNode.Addr)

	peerGen := genNode.PeerDb.Get(newNode.Addr)
	peerNew := newNode.PeerDb.Get(genNode.Addr)

	if peerGen == nil {
		t.Fatal("The genesis node does not contain the new node as a peer")
	}

	if peerNew == nil {
		t.Fatal("The new node does not contain the genesis node as a peer")
	}

	time.Sleep(1 * time.Second)

	ChkMnChnCons(t, []*pkg.Node{genNode, newNode})
	ChkNdPrs(t, genNode, []*pkg.Node{newNode})
	ChkNdPrs(t, newNode, []*pkg.Node{genNode})

	genNode.SendTx(100, 0, newNode.Id.GetPublicKeyBytes())
	genNode.SendTx(100, 0, genNode.Id.GetPublicKeyBytes())

	time.Sleep(6 * time.Second)

	newNode.StartMiner()

	ChkTxSeenLen(t, genNode, 1)
	ChkTxSeenLen(t, newNode, 1)

	time.Sleep(6 * time.Second)

	ChkMnChnCons(t, []*pkg.Node{genNode, newNode})

	AsrtBal(t, genNode, 100000)
	AsrtBal(t, newNode, 0)
}

func TestHndlChangedTxReq(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	newNode := pkg.New(pkg.DefaultConfig(GetFreePort()))

	genNode.Start()
	newNode.Start()
	genNode.ConnectToPeer(newNode.Addr)

	peerGen := genNode.PeerDb.Get(newNode.Addr)
	peerNew := newNode.PeerDb.Get(genNode.Addr)

	if peerGen == nil {
		t.Fatal("The genesis node does not contain the new node as a peer")
	}

	if peerNew == nil {
		t.Fatal("The new node does not contain the genesis node as a peer")
	}

	time.Sleep(1 * time.Second)

	ChkMnChnCons(t, []*pkg.Node{genNode, newNode})
	ChkNdPrs(t, genNode, []*pkg.Node{newNode})
	ChkNdPrs(t, newNode, []*pkg.Node{genNode})

	genNode.SendTx(100, 100, newNode.Id.GetPublicKeyBytes())
	newNode.SendTx(100, 100, genNode.Id.GetPublicKeyBytes())

	time.Sleep(6 * time.Second)

	newNode.StartMiner()

	ChkTxSeenLen(t, genNode, 1)
	ChkTxSeenLen(t, newNode, 1)

	time.Sleep(6 * time.Second)

	ChkMnChnCons(t, []*pkg.Node{genNode, newNode})

	AsrtBal(t, genNode, 99800)
	AsrtBal(t, newNode, 210)
}

func TestHndlTxReqNotEnough(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	newNode := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()
	genNode.ConnectToPeer(newNode.Addr)
	genNode.SendTx(200, 0, newNode.Id.GetPublicKeyBytes())

	time.Sleep(5 * time.Second)

	newNode.SendTx(100000000, 0, genNode.Id.GetPublicKeyBytes())

	ChkTxSeenLen(t, genNode, 0)
	ChkTxSeenLen(t, newNode, 0)
}

func TestAddHugeFee(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()

	genNode.Start()

	txPool := miner.NewTxPool(miner.DefaultConfig(-1))
	tx := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 100000000)

	txPool.Add(tx)

	if txPool.Ct.Load() != 2 {
		t.Errorf("Expected: 2 - Actual: %d", txPool.Ct.Load())
	}

	test := (*txPool.TxQ)[0]

	if test.P != miner.CalcPri(tx) {
		t.Errorf("Expected: 1 - Actual: %d", test.P)
	}
}

func TestAddZeroFee(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()

	genNode.Start()

	txPool := miner.NewTxPool(miner.DefaultConfig(-1))
	tx := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 5)

	txPool.Add(tx)

	if txPool.Ct.Load() != 2 {
		t.Errorf("Expected: 2 - Actual: %d", txPool.Ct.Load())
	}

	test := (*txPool.TxQ)[0]

	if test.P != 1 {
		t.Errorf("Expected: 1 - Actual: %d", test.P)
	}
}

func TestHndlMinerActive(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	genNode.Start()

	m := miner.New(genNode.Mnr.Conf, genNode.Id)
	m.Active.Store(true)

	tx := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 10)

	m.HndlTx(tx)

	if m.TxP.Ct.Load() != 1 {
		t.Errorf("Expected: %v - Actual;: %v", 1, m.TxP.Ct.Load())
	}
}

func TestHndlMinerInactive(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()

	genNode.Start()

	m := miner.New(genNode.Mnr.Conf, genNode.Id)
	m.Active.Store(false)

	tx := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 10)

	m.HndlTx(tx)

	poolUpdated := <-m.PoolUpdated

	if poolUpdated {
		t.Errorf("Pool is updated")
	}
}

func TestCalcNil(t *testing.T) {
	priority := miner.CalcPri(nil)

	if priority != 0 {
		t.Errorf("Expected: 0 - Actual: %d", priority)
	}
}

func TestCalcPri(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	genNode.Start()

	tx := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 2)

	priority := miner.CalcPri(tx)

	if priority != (tx.SumInputs()-tx.SumOutputs())*100.0/tx.Sz() {
		t.Errorf("Expected: %d - Actual: %d", (tx.SumInputs()-tx.SumOutputs())*100.0/tx.Sz(), priority)
	}
}

func TestHndlZeroFeesGenCB(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()

	inUTXO1 := &txo.TransactionOutput{
		Amount:        0,
		LockingScript: hex.EncodeToString(genNode.Id.GetPublicKeyBytes()),
	}

	sig1, _ := inUTXO1.MkSig(genNode.Id)

	TxI1 := &txi.TransactionInput{
		TransactionHash: inUTXO1.Hash(),
		OutputIndex:     0,
		UnlockingScript: sig1,
		Amount:          inUTXO1.Amount,
	}

	inUTXO2 := &txo.TransactionOutput{
		Amount:        0,
		LockingScript: hex.EncodeToString(genNode.Id.GetPublicKeyBytes()),
	}

	sig2, _ := inUTXO2.MkSig(genNode.Id)

	TxI2 := &txi.TransactionInput{
		TransactionHash: inUTXO2.Hash(),
		OutputIndex:     1,
		UnlockingScript: sig2,
		Amount:          inUTXO2.Amount,
	}

	TxO1 := &txo.TransactionOutput{
		Amount:        0,
		LockingScript: hex.EncodeToString(genNode.Id.GetPublicKeyBytes()),
	}
	TxO2 := &txo.TransactionOutput{
		Amount:        0,
		LockingScript: hex.EncodeToString(genNode.Id.GetPublicKeyBytes()),
	}

	transaction := &tx.Transaction{
		Version:  genNode.Wallet.Conf.TxVer,
		Inputs:   []*txi.TransactionInput{TxI1, TxI2},
		Outputs:  []*txo.TransactionOutput{TxO1, TxO2},
		LockTime: genNode.Wallet.Conf.DefLckTm,
	}

	tx := genNode.Mnr.GenCBTx([]*tx.Transaction{transaction})

	if tx.SumOutputs() != genNode.Mnr.Conf.InitSubsdy {
		t.Errorf("Outputs not equal to initial")
	}
}

func TestHndlLowFeesGenCB(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()

	transaction := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 0)

	CBTX := genNode.Mnr.GenCBTx([]*tx.Transaction{transaction})

	if CBTX.SumOutputs() > genNode.Mnr.Conf.InitSubsdy {
		t.Errorf("Outputs are larger than the initial")
	}
}

func TestHndlNilGenCBArray(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()

	transaction := CreateTx(genNode, genNode.Id.GetPublicKeyBytes(), 2)

	txArray := genNode.Mnr.GenCBTx([]*tx.Transaction{nil, transaction})

	if txArray != nil {
		t.Errorf("Array is nil")
	}
}

func TestHndlTxNil(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()

	genNode.Start()

	mnr := miner.New(genNode.Mnr.Conf, genNode.Id)

	mnr.Active.Store(true)
	mnr.HndlTx(nil)

	if mnr.TxP.Ct.Load() != 1 {
		t.Errorf("Expected: %v - Actual: %v", 0, mnr.TxP.Ct.Load())
	}
}

func TestHndlZeroTxGenCB(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()

	tx := genNode.Mnr.GenCBTx([]*tx.Transaction{})

	if tx != nil {
		t.Errorf("No transaction")
	}
}

func TestHndlNilTxGenCB(t *testing.T) {
	utils.SetDebug(true)

	genNode := NewGenNd()
	genNode.Conf.MnrConf.InitPOWD = utils.CalcPOWD(1)

	genNode.Start()
	genNode.StartMiner()

	tx := genNode.Mnr.GenCBTx(nil)

	if tx != nil {
		t.Errorf("No transaction")
	}
}
