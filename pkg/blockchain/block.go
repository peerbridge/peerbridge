package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type SHA256 [sha256.Size]byte

type Block struct {
	Index        string        `json:"index" pg:"type:uuid,default:gen_random_uuid(),pk,unique,notnull"` // random uuid primary key
	Timestamp    time.Time     `json:"timestamp" pg:"default:now(),notnull"`
	Transactions []Transaction `json:"transactions" pg:"rel:has-many"`
	ParentIndex  string        `json:"parentIndex"`
}

// TODO: include Hash into Block
func (b *Block) Hash() SHA256 {
	jsonBytes, _ := json.Marshal(b)
	digest := sha256.Sum256(jsonBytes)
	return digest
}
