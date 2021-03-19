package secp256k1

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	// Use the ethereum implementation of the secp256k1
	// elliptic curve digital signature algorithm, which
	// bridges to the C-implementation of Bitcoin
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	// The compressed byte length of a secp256k1 public key.
	PublicKeyByteLength = 33
)

type PublicKey struct {
	// The key bytes, in their compressed form.
	CompressedBytes [PublicKeyByteLength]byte `json:"bytes"`
}

func (p *PublicKey) Short() (result [3]byte) {
	copy(result[:], p.CompressedBytes[:3])
	return result
}

func (p *PublicKey) Decompress() (*[]byte, error) {
	x, y := ethsecp256k1.DecompressPubkey(p.CompressedBytes[:])
	if x == nil || y == nil {
		return nil, errors.New("Invalid compressed public key!")
	}
	decompressedBytes := ethsecp256k1.S256().Marshal(x, y)
	return &decompressedBytes, nil
}

func (p *PublicKey) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(p.CompressedBytes[:])
	return json.Marshal(hexString)
}

func (p *PublicKey) UnmarshalJSON(data []byte) error {
	var hexString string
	err := json.Unmarshal(data, &hexString)
	if err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(bytes) != PublicKeyByteLength {
		return errors.New("Invalid secp256k1 public key byte length!")
	}
	var fixedBytes [PublicKeyByteLength]byte
	copy(fixedBytes[:], bytes[:PublicKeyByteLength])
	p.CompressedBytes = fixedBytes
	return nil
}
