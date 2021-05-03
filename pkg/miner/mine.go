package miner

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"context"
	"encoding/hex"
	"math"
)

// Mine waits to be told to mine a block
// or to kill it's thread. If it is asked
// to mine, it selects the transactions
// with the highest priority to add to the
// mining pool. The nonce is then attempted
// to be found unless the miner is stopped.
func (m *Miner) Mine() {
	ctx, cancel := context.WithCancel(context.Background())
	for {
		<-m.PoolUpdated
		cancel()
		if !m.Active.Load() {
			continue
		}
		ctx, cancel = context.WithCancel(context.Background())
		go func(ctx context.Context) {
			if !m.TxP.PriMet() {
				return
			}
			m.Mining.Store(true)
			m.MiningPool = m.NewMiningPool()
			txs := append([]*tx.Transaction{m.GenCBTx(m.MiningPool)}, m.MiningPool...)
			b := block.New(m.PrvHsh, txs, m.DifTrg())
			result := m.CalcNonce(ctx, b)
			m.Mining.Store(false)
			if result {
				utils.Debug.Printf("%v mined %v %v", utils.FmtAddr(m.Addr), b.NameTag(), b.Summarize())
				m.SendBlk <- b
				m.HndlBlk(b)
			}
		}(ctx)
	}
	cancel()
}

// Returns boolean to indicate success
func (m *Miner) CalcNonce(ctx context.Context, b *block.Block) bool {
	for i := uint32(0); i < m.Conf.NncLim; i++ {
		select {
		case <-ctx.Done():
			return false
		default:
			b.Hdr.Nonce = i
			if b.SatisfiesPOW(m.DifTrg()) {
				return true
			}
		}
	}
	return false
}

// DifTrg (DifficultyTarget) calculates the
// current difficulty target.
// Returns:
// string the difficulty target as a hex
// string
func (m *Miner) DifTrg() string {
	return m.Conf.InitPOWD
}

// GenCBTx (GenerateCoinbaseTransaction) generates a coinbase
// transaction based off the transactions in the mining pool.
// It does this by adding the fee reward to the minting reward.
// Inputs:
// txs	[]*tx.Transaction the transactions (besides the
// coinbase tx) that the miner is mining to a block
// Returns:
// the coinbase transaction that pays the miner the reward
// for mining the block
func (m *Miner) GenCBTx(txs []*tx.Transaction) *tx.Transaction {
	if len(txs) <= 0 || txs == nil {
		return nil
	}

	var inputs uint32
	var outputs uint32
	var mintingReward uint32

	for i := range txs {
		if txs[i] == nil {
			return nil
		}
		inputs += txs[i].SumInputs()
		outputs += txs[i].SumOutputs()
	}

	fee := inputs - outputs

	divisor := m.ChnLen.Load() / m.Conf.SubsdyHlvRt

	if divisor > m.Conf.MxHlvgs {
		mintingReward = 0
	} else {
		mintingReward = m.Conf.InitSubsdy / uint32((math.Pow(2, float64(divisor))))
	}

	reward := mintingReward + fee
	PCBTxO := []*proto.TransactionOutput{proto.NewTxOutpt(reward, hex.EncodeToString(m.Id.GetPublicKeyBytes()))}
	PCBTx := proto.NewTx(m.Conf.Ver, []*proto.TransactionInput{}, PCBTxO, m.Conf.DefLckTm)
	CBTx := tx.Deserialize(PCBTx)

	return CBTx
}
