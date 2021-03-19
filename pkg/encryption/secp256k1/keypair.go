package secp256k1

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"

	// Use the ethereum implementation of the secp256k1
	// elliptic curve digital signature algorithm, which
	// bridges to the C-implementation of Bitcoin
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type KeyPair struct {
	// The public key of the key pair, in its compressed form.
	PublicKey PublicKey `json:"publicKey"`
	// The private key of the key pair.
	PrivateKey PrivateKey `json:"privateKey"`
}

func GenerateNewKeyPair(keypath string) (*KeyPair, error) {
	key, err := ecdsa.GenerateKey(ethsecp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	var publicKeyBytes [PublicKeyByteLength]byte
	copy(publicKeyBytes[:], ethsecp256k1.CompressPubkey(key.X, key.Y))
	publicKey := PublicKey{
		CompressedBytes: publicKeyBytes,
	}

	var privateKeyBytes [PrivateKeyByteLength]byte
	blob := key.D.Bytes()
	copy(privateKeyBytes[PrivateKeyByteLength-len(blob):], blob)
	privateKey := PrivateKey{
		Bytes: privateKeyBytes,
	}

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

func LoadKeyPair(keypath string) (*KeyPair, error) {
	bytes, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}
	var keyPair KeyPair
	err = json.Unmarshal(bytes, &keyPair)
	if err != nil {
		return nil, err
	}
	return &keyPair, nil
}

func StoreNewKeyPair(keypath string) (*KeyPair, error) {
	keyPair, err := GenerateNewKeyPair(keypath)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(keyPair)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(keypath, bytes, 0644)
	if err != nil {
		return nil, err
	}
	return keyPair, nil
}
