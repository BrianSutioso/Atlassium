package pkg

import (
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/id"
	"BrunoCoin/pkg/miner"
	"BrunoCoin/pkg/wallet"
	"time"
)

// Config is the configuration for the node.
// IdConf is the configuration for the id,
// MnrConf is the configuration for the miner,
// WtConf is the configuration for the wallet,
// ChainConf is the configuration for the blockchain,
// Version is the version that the node is (used for
// software updates),
// PeerLimit is the maximum amount of peers the node
// is allowed to have,
// AddrLimit is the maximum amount of addresses the
// node is allowed to keep track of.
// Port is the port that the node should run on,
// MxBlkSz is the maximum allowed block size,
type Config struct {
	IdConf    *id.Config
	MnrConf   *miner.Config
	WtConf    *wallet.Config
	ChainConf *blockchain.Config

	CstmID    bool
	CstmIDObj id.ID

	Version    int
	PeerLimit  int
	AddrLimit  int
	Port       int
	VerTimeout time.Duration

	MxBlkSz uint32
}

// DefaultConfig creates a Config object that
// contains basic/standard configurations for
// the node. To do this, it also calls the default
// config methods for the other more specific
// configs (such as configs for id, miner, wallet,
// and Chain).
// Inputs:
// port int the port that the node should start
// on
func DefaultConfig(port int) *Config {
	c := &Config{
		IdConf:     id.DefaultConfig(),
		MnrConf:    miner.DefaultConfig(-1),
		WtConf:     wallet.DefaultConfig(),
		ChainConf:  blockchain.DefaultConfig(),
		Version:    0,
		PeerLimit:  20,
		AddrLimit:  1000,
		Port:       port,
		VerTimeout: time.Second * 2,
		MxBlkSz:    10000000,
	}
	return c
}

func TestingConfig(port int) *Config {
	c := &Config{
		IdConf:     id.DefaultConfig(),
		MnrConf:    miner.DefaultConfig(-1),
		WtConf:     wallet.DefaultConfig(),
		ChainConf:  blockchain.DefaultConfig(),
		Version:    0,
		PeerLimit:  20,
		AddrLimit:  1000,
		Port:       port,
		VerTimeout: time.Second * 2,
		MxBlkSz:    10000000,
	}
	return c
}

// NilConfig has an ID, but it doesn't have any
// other functionality other than an RPC server.
// Inputs:
// port int the port that the node should start
// on
func NilConfig(port int) *Config {
	return &Config{
		IdConf:     id.DefaultConfig(),
		MnrConf:    miner.NilConfig(-1),
		WtConf:     wallet.NilConfig(),
		ChainConf:  blockchain.NilConfig(),
		Version:    0,
		PeerLimit:  20,
		AddrLimit:  1000,
		Port:       port,
		VerTimeout: time.Second * 2,
		MxBlkSz:    10000000,
	}
}

// NoMnrConfig is a configuration with default
// settings except that there is no miner
// Inputs:
// port int the port that the node should start
// on
func NoMnrConfig(port int) *Config {
	return &Config{
		IdConf:     id.DefaultConfig(),
		MnrConf:    miner.NilConfig(-1),
		WtConf:     wallet.DefaultConfig(),
		ChainConf:  blockchain.DefaultConfig(),
		Version:    1,
		PeerLimit:  20,
		AddrLimit:  1000,
		Port:       port,
		VerTimeout: time.Second * 2,
		MxBlkSz:    10000000,
	}
}

// SmallTxPConfig is a configuration with default
// settings except that the txpool cap is very small
// Inputs:
// port int the port that the node should start
// on
func SmallTxPConfig(port int) *Config {
	c := &Config{
		IdConf:     id.DefaultConfig(),
		MnrConf:    miner.SmallTxPCapConfig(-1),
		WtConf:     wallet.DefaultConfig(),
		ChainConf:  blockchain.DefaultConfig(),
		Version:    0,
		PeerLimit:  20,
		AddrLimit:  1000,
		Port:       port,
		VerTimeout: time.Second * 2,
		MxBlkSz:    10000000,
	}
	return c
}
