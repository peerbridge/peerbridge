package secp256k1

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

const (
	// The compressed byte length of a secp256k1 public key.
	PublicKeyByteLength = 33
)

type PublicKey struct {
	Bytes [PublicKeyByteLength]byte `json:"bytes"`
}

func (p *PublicKey) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(p.Bytes[:])
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
	p.Bytes = fixedBytes
	return nil
}
