package wallet

// Config represents the configuration (settings)
// for the wallet.
// HasWt (HasWallet) defines whether the wallet
// should even exist on the node.
// TxRplyThresh (TransactionReplayThreshold)
// defines the time (represented by blocks seen)
// in which the wallet will resend a transaction
// to the network.
// SafeBlkAmt (SafeBlockAmount) defines the amount
// of blocks that need to be on top of the block
// that contains a transaction for that transaction
// to be considered valid by the wallet.
// TxVer (TransactionVersion) is the same as the
// software version of the node.
// DefLckTm (DefaultLockTime) is the default lock
// time (when the utxo can be spent)
type Config struct {
	HasWt        bool
	TxRplyThresh uint32
	SafeBlkAmt   int
	TxVer        uint32
	DefLckTm     uint32
}

// DefaultConfig returns the standard/basic
// settings for the wallet
func DefaultConfig() *Config {
	return &Config{
		HasWt:        true,
		TxRplyThresh: 3,
		SafeBlkAmt:   5,
		TxVer:        0,
		DefLckTm:     0,
	}
}

// NilConfig returns settings that say
// the wallet should not exist.
func NilConfig() *Config {
	return &Config{
		HasWt:        false,
		TxRplyThresh: 0,
		SafeBlkAmt:   0,
		TxVer:        0,
		DefLckTm:     0,
	}
}
