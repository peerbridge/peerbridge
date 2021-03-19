package secp256k1

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

const (
	// The byte length of a secp256k1 private key.
	PrivateKeyByteLength = 32
)

type PrivateKey struct {
	Bytes [PrivateKeyByteLength]byte `json:"bytes"`
}

func (p *PrivateKey) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(p.Bytes[:])
	return json.Marshal(hexString)
}

func (p *PrivateKey) UnmarshalJSON(data []byte) error {
	var hexString string
	err := json.Unmarshal(data, &hexString)
	if err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(bytes) != PrivateKeyByteLength {
		return errors.New("Invalid secp256k1 private key byte length!")
	}
	var fixedBytes [PrivateKeyByteLength]byte
	copy(fixedBytes[:], bytes[:PrivateKeyByteLength])
	p.Bytes = fixedBytes
	return nil
}
