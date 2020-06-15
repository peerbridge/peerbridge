package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

const ISO_8601 = time.RFC3339

const AddressSize = 20

type SHA256 [sha256.Size]byte
type Address [AddressSize]byte
type TransactionData []byte

type Transaction struct {
	Sender   Address `json:"sender"`
	Receiver Address `json:"receiver"`
	Data     []byte  `json:"data"`
}

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

var BlockChain []Block
