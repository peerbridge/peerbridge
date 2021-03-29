package secp256k1

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"

	// Use the ethereum implementation of the secp256k1
	// elliptic curve digital signature algorithm, which
	// bridges to the C-implementation of Bitcoin
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var (
	ErrWrongPrivateKeyLength         = errors.New("Wrong private key length!")
	ErrPublicKeyReconstructionFailed = errors.New("Public key could not be reconstructed!")
)

type KeyPair struct {
	// The public key of the key pair, in its compressed form.
	PublicKey PublicKeyHexString `json:"publicKey"`
	// The private key of the key pair.
	PrivateKey PrivateKeyHexString `json:"privateKey"`
}

func GenerateNewKeyPair(keypath string) (*KeyPair, error) {
	key, err := ecdsa.GenerateKey(ethsecp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	var publicKeyBytes [PublicKeyByteLength]byte
	copy(publicKeyBytes[:], ethsecp256k1.CompressPubkey(key.X, key.Y))
	publicKeyHexString := hex.EncodeToString(publicKeyBytes[:])

	var privateKeyBytes [PrivateKeyByteLength]byte
	blob := key.D.Bytes()
	copy(privateKeyBytes[PrivateKeyByteLength-len(blob):], blob)
	privateKeyHexString := hex.EncodeToString(privateKeyBytes[:])

	return &KeyPair{
		PublicKey:  publicKeyHexString,
		PrivateKey: privateKeyHexString,
	}, nil
}

func LoadKeyPairFromPrivateKeyString(privateKeyHexString string) (*KeyPair, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHexString)
	if err != nil {
		return nil, err
	}
	if len(privateKeyBytes) != PrivateKeyByteLength {
		return nil, ErrWrongPrivateKeyLength
	}
	x, y := secp256k1.S256().ScalarBaseMult(privateKeyBytes)
	if x == nil || y == nil {
		return nil, ErrPublicKeyReconstructionFailed
	}
	var publicKeyBytes [PublicKeyByteLength]byte
	copy(publicKeyBytes[:], ethsecp256k1.CompressPubkey(x, y))
	publicKeyHexString := hex.EncodeToString(publicKeyBytes[:])

	return &KeyPair{publicKeyHexString, privateKeyHexString}, nil
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
