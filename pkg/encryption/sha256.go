package encryption

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"time"
)

const (
	SHA256ByteLength = 32
)

type SHA256HexString = string

func ZeroSHA256HexString() SHA256HexString {
	hashBytes := [SHA256ByteLength]byte{}
	hashString := hex.EncodeToString(hashBytes[:])
	return hashString
}

func RandomSHA256HexString() (*SHA256HexString, error) {
	hashBytes := &[SHA256ByteLength]byte{}
	rand.Seed(time.Now().UTC().UnixNano())
	_, err := rand.Read(hashBytes[:])
	if err != nil {
		return nil, err
	}
	hashString := hex.EncodeToString(hashBytes[:])
	return &hashString, nil
}

func SHA256HexStringToBytes(hexString string) (*[SHA256ByteLength]byte, error) {
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}
	if len(bytes) != SHA256ByteLength {
		return nil, errors.New("Invalid sha256 byte length!")
	}
	var fixedBytes [SHA256ByteLength]byte
	copy(fixedBytes[:], bytes[:SHA256ByteLength])
	return &fixedBytes, nil
}
