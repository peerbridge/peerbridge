package secp256k1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"

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
	signatureData, err := ethsecp256k1.Sign(input.Bytes[:], privateKeyBytes)
	if err != nil {
		return nil, err
	}
	if len(signatureData) != SignatureByteLength {
		return nil, ErrWrongSignatureLength
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
		return ErrSignatureNotVerifiable
	}
	return nil
}

// Get the signing input for a given object.
// To declare fields as to-be-signed, tag them with
// `sign:"yes"`. The algorithm will marshal the object's
// fields that are tagged in this way and use the bytes
// as an input for the signature computation.
//
// Example: suppose you have the struct type
// myAnonymousStruct := {
//     MyField1: string `json:"field1"` `sign:"yes"`
//	   MyField2: int `json:"field2"`
// }
// then the signing input will be a bytes array of the json:
// {"field1":"myValue"}
func GetSigningInput(object interface{}) (*SigningInput, error) {
	// Get all fields that are tagged with sign:"yes"
	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object)
	values := map[string]interface{}{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("sign") == "yes" {
			name := field.Tag.Get("json")
			values[name] = v.Field(i).Interface()
		}
	}
	// Marshal those fields to json and use
	// it to create the signing input
	bytes, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	input := NewSigningInput(bytes)
	return &input, nil
}

func ComputeSignature(object interface{}, p PrivateKeyHexString) (*SignatureHexString, error) {
	input, err := GetSigningInput(object)
	if err != nil {
		return nil, err
	}
	return input.Sign(p)
}

func VerifySignature(object interface{}, sig SignatureHexString) error {
	input, err := GetSigningInput(object)
	if err != nil {
		return err
	}
	return input.VerifySignature(sig)
}
