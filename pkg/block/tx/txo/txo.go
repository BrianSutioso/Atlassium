package txo

import (
	"BrunoCoin/pkg/id"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// TransactionOutput is a wrapper around the
// protobuf transaction output so methods
// can be called on it and additional fields
// may be added.
type TransactionOutput struct {
	Amount        uint32
	LockingScript string
	Liminal       bool
}

// IsUnlckd (IsUnlocked) tests whether an unlocking
// script successfully unlocks a locking script
// on the transaction output. It does this by using
// a public key, a message, and a signature to verify
// whether the signature is valid.
// Inputs:
// sig	string	signature a.k.a. unlocking script
// represented as a hex string.
// Returns:
// bool	true if the unlocking script actually
// unlocks the locking script. False otherwise.
func (o *TransactionOutput) IsUnlckd(sig string) bool {
	pkb, err := hex.DecodeString(o.LockingScript)
	if err != nil {
		fmt.Printf("ERROR {IsUnlckd}: Locking"+
			"script on the transaction {%v} couldn't"+
			"decode properly.\n", o.LockingScript)
		return false
	}
	pk, err := utils.Byt2PK(pkb)
	if err != nil {
		fmt.Printf("ERROR {IsUnlckd}:" +
			" utils.Byt2PK errored.\n")
		return false
	}
	h, err := hex.DecodeString(o.Hash())
	if err != nil {
		fmt.Printf("ERROR {IsUnlckd}: Could not"+
			" properly decode the hash {%v} of the"+
			" transaction output. ->"+
			" TransactionOutput.Hash() errored.\n", o.Hash())
		return false
	}
	sigB, err := hex.DecodeString(sig)
	if err != nil {
		fmt.Printf("ERROR {IsUnlckd}: Received "+
			"an unlocking script {%v} that couldn't "+
			"decode properly.\n", sig)
		return false
	}
	return ecdsa.VerifyASN1(pk, h, sigB)
}

// PrsTXOLoc (ParseTransactionOutputLocator) parses
// a "locator" for a transaction. A locator is a
// unique string that can identify every transaction
// output. They are represented in the form:
// "{transaction hash}-{index}", where transaction
// hash represents the hash of the transaction that
// the transaction output is within. The index
// represents the index in the outputs array on the
// transaction that would be the corresponding
// transaction output in question.
// Inputs:
// l	string	the locator of the transaction output.
// Returns:
// string	the transaction hash as a hex string
// uint32	the index into the outputs array
func PrsTXOLoc(l string) (string, uint32) {
	d := strings.Split(l, "-")
	i, err := strconv.ParseUint(d[1], 10, 32)
	if err != nil {
		fmt.Printf("ERROR {PrsTXOLoc}: Could not"+
			"parse the index out of the txo locator {%v}.\n", l)
		return "", 0
	}
	return d[0], uint32(i)
}

// MkTXOLoc (MakeTransactionOutputLocator) makes a
// transaction output locator. A locator is a
// unique string that can identify every transaction
// output. They are represented in the form:
// "{transaction hash}-{index}", where transaction
// hash represents the hash of the transaction that
// the transaction output is within. The index
// represents the index in the outputs array on the
// transaction that would be the corresponding
// transaction output in question.
// Inputs:
// h	string	the hash of the transaction
// as a hex string.
// i	uint32	the index into the outputs array.
// Returns:
// string	the locator for the transaction output.
func MkTXOLoc(h string, i uint32) string {
	return fmt.Sprintf("%v-%v", h, i)
}

// MkSig (MakeSignature) generates
// an unlocking script (a.k.a. signature) for the
// transaction output based on a private key.
// Inputs:
// id	id.ID	the id of the person wanting to
// unlock the particular transaction output.
// Returns:
// string	The signature represented as a hex string.
// error	Errors if the signature could not be
// produced or there was a decoding error.
func (o *TransactionOutput) MkSig(id id.ID) (string, error) {
	sk := id.GetPrivateKey()
	hB, err := hex.DecodeString(o.Hash())
	if err != nil {
		fmt.Printf("ERROR {TransactionOutput.MkSig}: "+
			"The hash of the transaction output {%v} could "+
			"not decode -> errored in "+
			"TransactionOutput.Hash().\n", o.Hash())
		return "", nil
	}
	sig, err := utils.Sign(sk, hB)
	if err != nil {
		fmt.Printf("ERROR {TransactionOutput.MkSig}: " +
			"The signature could not be formed.\n")
		return "", nil
	}
	return sig, nil
}

// Serialize serializes a transaction output
// into a protobuf transaction output so it
// can properly be sent over the network.
// Returns:
// *proto.TransactionOutput a pointer to a
// protobuf txo.
func (o *TransactionOutput) Serialize() *proto.TransactionOutput {
	return &proto.TransactionOutput{
		Amount:        o.Amount,
		LockingScript: o.LockingScript,
	}
}

// Deserialize deserializes a  protobuf transaction output
// into a regular transaction output so it
// can be properly tested
// Returns:
// *txo.TransactionOutput a pointer to a
// txo.

func Deserialize(ptx *proto.TransactionOutput) *TransactionOutput {
	op := &TransactionOutput{
		Amount:        ptx.Amount,
		LockingScript: ptx.LockingScript,
	}
	return op
}

// Hash hashes a transaction output.
// Returns:
// string	the hash of the transaction output
// represented as a hex string.
func (o *TransactionOutput) Hash() string {
	return utils.Hash([]byte(fmt.Sprintf("%v/%v", o.Amount, o.LockingScript)))
}
