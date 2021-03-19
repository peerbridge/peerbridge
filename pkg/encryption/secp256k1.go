package encryption

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"

	// Use the ethereum implementation of the secp256k1
	// elliptic curve digital signature algorithm, which
	// bridges to the C-implementation of Bitcoin
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	Secp256k1PublicKeyByteLength  = 65
	Secp256k1PrivateKeyByteLength = 32
)

type Secp256k1PublicKey struct {
	Bytes [Secp256k1PublicKeyByteLength]byte `json:"bytes"`
}

func (p *Secp256k1PublicKey) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(p.Bytes[:])
	return json.Marshal(hexString)
}

func (p *Secp256k1PublicKey) UnmarshalJSON(data []byte) error {
	var hexString string
	err := json.Unmarshal(data, &hexString)
	if err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(bytes) != Secp256k1PublicKeyByteLength {
		return errors.New("Invalid secp256k1 public key byte length!")
	}
	var fixedBytes [Secp256k1PublicKeyByteLength]byte
	copy(fixedBytes[:], bytes[:Secp256k1PublicKeyByteLength])
	p.Bytes = fixedBytes
	return nil
}

type Secp256k1PrivateKey struct {
	Bytes [Secp256k1PrivateKeyByteLength]byte `json:"bytes"`
}

func (p *Secp256k1PrivateKey) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(p.Bytes[:])
	return json.Marshal(hexString)
}

func (p *Secp256k1PrivateKey) UnmarshalJSON(data []byte) error {
	var hexString string
	err := json.Unmarshal(data, &hexString)
	if err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(bytes) != Secp256k1PrivateKeyByteLength {
		return errors.New("Invalid secp256k1 private key byte length!")
	}
	var fixedBytes [Secp256k1PrivateKeyByteLength]byte
	copy(fixedBytes[:], bytes[:Secp256k1PrivateKeyByteLength])
	p.Bytes = fixedBytes
	return nil
}

type Secp256k1KeyPair struct {
	PublicKey  Secp256k1PublicKey  `json:"publicKey"`
	PrivateKey Secp256k1PrivateKey `json:"privateKey"`
}

func GenerateNewSecp256k1KeyPair(keypath string) (*Secp256k1KeyPair, error) {
	key, err := ecdsa.GenerateKey(ethsecp256k1.S256(), rand.Reader)

	if err != nil {
		return nil, err
	}

	var publicKeyBytes [Secp256k1PublicKeyByteLength]byte
	copy(publicKeyBytes[:], elliptic.Marshal(ethsecp256k1.S256(), key.X, key.Y))
	publicKey := Secp256k1PublicKey{
		Bytes: publicKeyBytes,
	}

	var privateKeyBytes [Secp256k1PrivateKeyByteLength]byte
	blob := key.D.Bytes()
	copy(privateKeyBytes[Secp256k1PrivateKeyByteLength-len(blob):], blob)
	privateKey := Secp256k1PrivateKey{
		Bytes: privateKeyBytes,
	}

	return &Secp256k1KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

func LoadSecp256k1KeyPair(keypath string) (*Secp256k1KeyPair, error) {
	bytes, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}
	var keyPair Secp256k1KeyPair
	err = json.Unmarshal(bytes, &keyPair)
	if err != nil {
		return nil, err
	}
	return &keyPair, nil
}

func StoreNewSecp256k1KeyPair(keypath string) (*Secp256k1KeyPair, error) {
	keyPair, err := GenerateNewSecp256k1KeyPair(keypath)
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
