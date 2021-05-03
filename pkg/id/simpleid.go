package id

import (
	"BrunoCoin/pkg/utils"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
)

type SimpleID struct {
	PrivateKey      *ecdsa.PrivateKey
	PrivateKeyBytes []byte
	PublicKey       *ecdsa.PublicKey
	PublicKeyBytes  []byte
}

func LoadInSmplID(pubK string, privK string) (*SimpleID, error) {
	pkB, _ := hex.DecodeString(pubK)
	skB, _ := hex.DecodeString(privK)
	pk, _ := utils.Byt2PK(pkB)
	pVk, _ := utils.Byt2SK(skB)
	id := &SimpleID{
		PrivateKey: pVk,
		PrivateKeyBytes: pkB,
		PublicKey:  pk,
		PublicKeyBytes: pkB,
	}
	return id, nil
}

func CreateSimpleID() (*SimpleID, error) {
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	id := &SimpleID{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
	}
	privateKeyBytes, err1 := id.PrivateKeyToBytes(id.PrivateKey)
	publicKeyBytes, err2 := id.PublicKeyToBytes(id.PublicKey)
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	id.PrivateKeyBytes = privateKeyBytes
	id.PublicKeyBytes = publicKeyBytes
	return id, nil
}

func (id *SimpleID) GetPrivateKey() *ecdsa.PrivateKey {
	return id.PrivateKey
}

func (id *SimpleID) GetPrivateKeyBytes() []byte {
	return id.PrivateKeyBytes
}

func (id *SimpleID) GetPublicKey() *ecdsa.PublicKey {
	return id.PublicKey
}

func (id *SimpleID) GetPublicKeyBytes() []byte {
	return id.PublicKeyBytes
}

func (id *SimpleID) BytesToPublicKey(bytes []byte) (*ecdsa.PublicKey, error) {
	genericPublicKey, err := x509.ParsePKIXPublicKey(bytes)
	if err != nil {
		return nil, err
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey, nil
}

func BytesToPublicKey(bytes []byte) (*ecdsa.PublicKey, error) {
	genericPublicKey, err := x509.ParsePKIXPublicKey(bytes)
	if err != nil {
		return nil, err
	}
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey, nil
}

func (id *SimpleID) PublicKeyToBytes(key *ecdsa.PublicKey) ([]byte, error) {
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(key)
	return x509EncodedPub, err
}

func (id *SimpleID) BytesToPrivateKey(bytes []byte) (*ecdsa.PrivateKey, error) {
	privateKey, err := x509.ParseECPrivateKey(bytes)
	return privateKey, err
}

func (id *SimpleID) PrivateKeyToBytes(key *ecdsa.PrivateKey) ([]byte, error) {
	x509Encoded, err := x509.MarshalECPrivateKey(key)
	return x509Encoded, err
}
