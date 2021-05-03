package id

import (
	"crypto/ecdsa"
)

type ID interface {
	GetPrivateKey() *ecdsa.PrivateKey
	GetPrivateKeyBytes() []byte
	GetPublicKey() *ecdsa.PublicKey
	GetPublicKeyBytes() []byte
	BytesToPublicKey(bytes []byte) (*ecdsa.PublicKey, error)
	PublicKeyToBytes(key *ecdsa.PublicKey) ([]byte, error)
	BytesToPrivateKey(bytes []byte) (*ecdsa.PrivateKey, error)
	PrivateKeyToBytes(key *ecdsa.PrivateKey) ([]byte, error)
}

func New(conf *Config) (ID, error) {
	return CreateSimpleID()
}
