package secp256k1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	// Use the ethereum implementation of the secp256k1
	// elliptic curve digital signature algorithm, which
	// bridges to the C-implementation of Bitcoin
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	// The length of a secp256k1 signature.
	SignatureByteLength = 65
	// The length of the secp256k1 signature input.
	SigningInputLength = 32
)

type Signature struct {
	Bytes [SignatureByteLength]byte
}

func (s *Signature) Equals(other *Signature) bool {
	return bytes.Compare(s.Bytes[:], other.Bytes[:]) == 0
}

func (s *Signature) Short() (result [3]byte) {
	copy(result[:], s.Bytes[:3])
	return result
}

func (s *Signature) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(s.Bytes[:])
	return json.Marshal(hexString)
}

func (s *Signature) UnmarshalJSON(data []byte) error {
	var hexString string
	err := json.Unmarshal(data, &hexString)
	if err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(bytes) != SignatureByteLength {
		return errors.New("Invalid secp256k1 signature byte length!")
	}
	var fixedBytes [SignatureByteLength]byte
	copy(fixedBytes[:], bytes[:SignatureByteLength])
	s.Bytes = fixedBytes
	return nil
}

type SigningInput struct {
	Bytes [SigningInputLength]byte
}

// Create a new signing input.
//
// The secp256k1 signature algorithm takes a 32 byte
// input vector, so we hash the data and use the sha256
// hash (which is 32 byte long) to generate the signature.
func NewSigningInput(data []byte) (input SigningInput) {
	hasher := sha256.New()
	hasher.Write(data)
	copy(input.Bytes[:], hasher.Sum(nil)[:SigningInputLength])
	return input
}

func (input *SigningInput) Sign(p *PrivateKey) (*Signature, error) {
	signatureData, err := ethsecp256k1.Sign(input.Bytes[:], p.Bytes[:])
	if err != nil {
		return nil, err
	}
	if len(signatureData) != SignatureByteLength {
		return nil, errors.New("Wrong signature length!")
	}
	var signatureFixedBytes [SignatureByteLength]byte
	copy(signatureFixedBytes[:], signatureData[:SignatureByteLength])
	return &Signature{signatureFixedBytes}, nil
}

func (input *SigningInput) VerifySignature(s *Signature, p *PublicKey) error {
	// The verification takes 64 bytes as an input.
	// This is, our signature (R, S, V) without the V byte.
	signatureWithoutV := s.Bytes[:len(s.Bytes)-1]
	if len(signatureWithoutV) != 64 {
		panic("The signature should always have length 65!")
	}
	// The public key is in a compressed format,
	// so we will need to decompress it
	decompressedP, err := p.Decompress()
	if err != nil {
		return err
	}

	isVerified := ethsecp256k1.VerifySignature(
		*decompressedP, input.Bytes[:], signatureWithoutV,
	)
	if !isVerified {
		return errors.New("Signature could not be verified!")
	}
	return nil
}
