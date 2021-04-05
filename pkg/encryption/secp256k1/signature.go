package secp256k1

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"

	// Use the ethereum implementation of the secp256k1
	// elliptic curve digital signature algorithm, which
	// bridges to the C-implementation of Bitcoin
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

const (
	// The length of a secp256k1 signature.
	SignatureByteLength = 64
	// The length of the secp256k1 signature input.
	SigningInputLength = 32
)

var (
	ErrWrongSignatureLength   = errors.New("Wrong signature length!")
	ErrSignatureNotVerifiable = errors.New("Signature could not be verified!")
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
	// Signature data is 64 bytes (including a recovery id)
	signatureData, err := ethsecp256k1.Sign(input.Bytes[:], privateKeyBytes)
	if err != nil {
		return nil, err
	}
	if len(signatureData) != (SignatureByteLength + 1) {
		return nil, ErrWrongSignatureLength
	}
	var signatureFixedBytes [SignatureByteLength]byte
	copy(signatureFixedBytes[:], signatureData[:SignatureByteLength])
	signatureString := hex.EncodeToString(signatureFixedBytes[:])
	return &signatureString, nil
}

func (input *SigningInput) VerifySignature(s SignatureHexString, sender PublicKeyHexString) error {
	signatureBytes, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	if len(signatureBytes) != SignatureByteLength {
		panic("The signature should always have length 64!")
	}

	senderBytes, err := hex.DecodeString(sender)
	if err != nil {
		return err
	}

	isVerified := ethsecp256k1.VerifySignature(
		senderBytes, input.Bytes[:], signatureBytes,
	)
	if !isVerified {
		return ErrSignatureNotVerifiable
	}
	return nil
}

type Signable interface {
	GetSignString() string
	GetSender() PublicKeyHexString
}

func ComputeSignature(s Signable, p PrivateKeyHexString) (*SignatureHexString, error) {
	str := s.GetSignString()
	log.Println(str)
	data := []byte(str)
	input := NewSigningInput(data)
	return input.Sign(p)
}

func VerifySignature(s Signable, sig SignatureHexString) error {
	str := s.GetSignString()
	log.Println(str)
	data := []byte(str)
	input := NewSigningInput(data)
	return input.VerifySignature(sig, s.GetSender())
}
