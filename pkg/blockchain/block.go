package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type SHA256 [sha256.Size]byte

type Block struct {
	Index        uint64        `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	ParentHash   SHA256        `json:"parentHash"`
	Transactions []Transaction `json:"transactions"`
}

func (b *Block) Hash() SHA256 {
	jsonBytes, _ := json.Marshal(b)
	digest := sha256.Sum256(jsonBytes)
	return digest
}
