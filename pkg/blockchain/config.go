package blockchain

// GENPK is the public key that was used
// for the genesis transaction on the
// genesis block.
var GENPK = "3059301306072a8648ce3d020106082a8648ce3d030107034200042418a20458559ae13a0d4bb6ac284c66a5cebb5689563d4cf573473d8c6d5abfa9a21a65dbb3ba2f2d930be7f763f940f9864abaf199a0f0d8d14bedda2dcad9"

// GENPVK is the public key that was used
// for the genesis transaction on the
// genesis block.
var GENPVK = "307702010104202456b0e8bed5c27dcadb044df1af8eaf714084b61a23d17359fb09f3c3f5fff5a00a06082a8648ce3d030107a144034200042418a20458559ae13a0d4bb6ac284c66a5cebb5689563d4cf573473d8c6d5abfa9a21a65dbb3ba2f2d930be7f763f940f9864abaf199a0f0d8d14bedda2dcad9"

// Config represents the settings for the
// blockchain.
// HasChn True if the node wants to store
// a copy of the blockchain.
// InitSbsdy is the amount of money given
// to GenPK in the genesis transaction.
// GenPK is the public key for the genesis
// transaction.
type Config struct {
	HasChn    bool
	InitSbsdy uint32
	GenPK     string
}

// DefaultConfig returns the default
// settings for the configuration of the
// blockchain.
func DefaultConfig() *Config {
	return &Config{
		HasChn:    true,
		InitSbsdy: 100000,
		GenPK:     GENPK,
	}
}

// NilConfig returns a config that sets
// HasChn to false, meaning the node will
// not have a copy of the blockchain.
func NilConfig() *Config {
	return &Config{
		HasChn:    false,
		InitSbsdy: 100000,
		GenPK:     GENPK,
	}
}
