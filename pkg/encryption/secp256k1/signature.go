package secp256k1

import (
	"crypto/sha256"
	"encoding/hex"
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

type SignatureHexString = string

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

func (input *SigningInput) Sign(p PrivateKeyHexString) (*SignatureHexString, error) {
	privateKeyBytes, err := hex.DecodeString(p)
	if err != nil {
		return nil, err
	}
	signatureData, err := ethsecp256k1.Sign(input.Bytes[:], privateKeyBytes)
	if err != nil {
		return nil, err
	}
	if len(signatureData) != SignatureByteLength {
		return nil, errors.New("Wrong signature length!")
	}
	var signatureFixedBytes [SignatureByteLength]byte
	copy(signatureFixedBytes[:], signatureData[:SignatureByteLength])
	signatureString := hex.EncodeToString(signatureFixedBytes[:])
	return &signatureString, nil
}

func (input *SigningInput) VerifySignature(s SignatureHexString) error {
	signatureBytes, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(signatureBytes) != SignatureByteLength {
		panic("The signature should always have length 65!")
	}
	// The verification takes 64 bytes as an input.
	// This is, our signature (R, S, V) without the V byte.
	signatureWithoutV := signatureBytes[:len(signatureBytes)-1]

	// Recover the public key from the signature
	pubkey, err := ethsecp256k1.RecoverPubkey(input.Bytes[:], signatureBytes)
	if err != nil {
		return err
	}

	isVerified := ethsecp256k1.VerifySignature(
		pubkey, input.Bytes[:], signatureWithoutV,
	)
	if !isVerified {
		return errors.New("Signature could not be verified!")
	}
	return nil
}
