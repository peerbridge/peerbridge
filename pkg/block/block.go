package block

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

const ISO_8601 = time.RFC3339

type Block struct {
	Index     uint64
	Timestamp time.Time
	PrevHash  string
}

func (b *Block) Hash() string {
	record := strconv.FormatUint(b.Index, 10) + b.Timestamp.Format(ISO_8601) + b.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}
