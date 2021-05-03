package miner

import (
	"BrunoCoin/pkg/utils"
	"math"
)

// Config represents the settings for the
// Miner.
// HasMnr (HasMiner) defines whether or not
// the node will have a miner.
// Ver (Ver) defines the software version
// of the node.
// DefLckTm (DefineLockTime) defines the lock
// time that should be on the coinbase transaction.
// TxPCap defines the maximum number of
// transactions allowed in the transaction pool.
// PriLim defines the priority threshold that
// must be met for the miner to start mining a
// group of transactions
// BlkSz defines the maximum size a block can be.
// NncLim defines the maximum nonce that miners
// are willing to mine to.
// InitSubsdy defines the initial subsidy given
// to miners for the minting reward before any
// havlings.
// SubsdyHlvRt defines the rate (in terms of
// blocks added) at which the initial subsidy
// is havled.
// MxHlvgs defines the maximum number of
// halvings that are allowed before the subsidy
// becomes 0.
// InitPOWD represents the inital proof of
// work difficulty. This is helpful in the
// config because now some nodes can be toggled
// to have a higher proof of work than others,
// which is essentially adjusting the speeds of miners
// on the network.
type Config struct {
	HasMnr bool

	Ver      uint32
	DefLckTm uint32

	TxPCap uint32
	PriLim uint32

	BlkSz  uint32
	NncLim uint32

	InitSubsdy  uint32
	SubsdyHlvRt uint32
	MxHlvgs     uint32
	InitPOWD    string
}

// DefaultConfig returns the default settings
// for the Miner.
func DefaultConfig(powdNumZeros int) *Config {
	return &Config{
		HasMnr:      true,
		Ver:         0,
		DefLckTm:    0,
		TxPCap:      50,
		PriLim:      10,
		BlkSz:       1000,
		NncLim:      uint32(math.Pow(2, 20)),
		InitSubsdy:  10,
		SubsdyHlvRt: 10,
		MxHlvgs:     10,
		InitPOWD:    utils.CalcPOWD(powdNumZeros),
	}
}

// NilConfig returns the settings for a
// miner where HasMnr is set to false,
// so that the node won't support the
// mining functionality.
func NilConfig(powdNumZeros int) *Config {
	return &Config{
		HasMnr:      false,
		Ver:         0,
		DefLckTm:    0,
		TxPCap:      50,
		PriLim:      10,
		BlkSz:       1000,
		NncLim:      uint32(math.Pow(2, 20)),
		InitSubsdy:  10,
		SubsdyHlvRt: 10,
		MxHlvgs:     10,
		InitPOWD:    utils.CalcPOWD(powdNumZeros),
	}
}

// SmallTxPCapConfig returns the settings for a
// a miner where everything is the default,
// except for a very small txpool cap
func SmallTxPCapConfig(powdNumZeros int) *Config {
	return &Config{
		HasMnr:      true,
		Ver:         0,
		DefLckTm:    0,
		TxPCap:      1,
		PriLim:      10,
		BlkSz:       1000,
		NncLim:      uint32(math.Pow(2, 20)),
		InitSubsdy:  10,
		SubsdyHlvRt: 10,
		MxHlvgs:     10,
		InitPOWD:    utils.CalcPOWD(powdNumZeros),
	}
}
