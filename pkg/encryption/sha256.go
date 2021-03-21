package encryption

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/rand"
	"time"
)

const (
	SHA256ByteLength = 32
)

type SHA256 struct {
	Bytes [SHA256ByteLength]byte
}

func (h *SHA256) Equals(other *SHA256) bool {
	return bytes.Compare(h.Bytes[:], other.Bytes[:]) == 0
}

func RandomSHA256() (*SHA256, error) {
	hash := &SHA256{}
	rand.Seed(time.Now().UTC().UnixNano())
	_, err := rand.Read(hash.Bytes[:])
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (h *SHA256) Short() (result [3]byte) {
	copy(result[:], h.Bytes[:3])
	return result
}

func (h *SHA256) MarshalJSON() ([]byte, error) {
	hexString := hex.EncodeToString(h.Bytes[:])
	return json.Marshal(hexString)
}

func (h *SHA256) UnmarshalJSON(data []byte) error {
	var hexString string
	err := json.Unmarshal(data, &hexString)
	if err != nil {
		return err
	}
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}
	if len(bytes) != SHA256ByteLength {
		return errors.New("Invalid sha256 byte length!")
	}
	var fixedBytes [SHA256ByteLength]byte
	copy(fixedBytes[:], bytes[:SHA256ByteLength])
	h.Bytes = fixedBytes
	return nil
}
