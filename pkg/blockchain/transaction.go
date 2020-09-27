package blockchain

import (
	"crypto/rsa"
	"time"

	"github.com/peerbridge/peerbridge/pkg/encryption"
)

type PEMPublicKey = string

type Transaction struct {
	Index      string       `json:"index" pg:"type:uuid,default:gen_random_uuid(),pk,unique,notnull"` // random uuid primary key
	Sender     PEMPublicKey `json:"sender" pg:",notnull"`
	Receiver   PEMPublicKey `json:"receiver" pg:",notnull"`
	Timestamp  time.Time    `json:"timestamp" pg:"default:now(),notnull"`
	Data       []byte       `json:"data" pg:",notnull"` // combined enrypted message of nonce|message|tag as returned by SealedBox.combined
	BlockIndex string       `json:"blockIndex" pg:"type:uuid"`
}

func (t *Transaction) SenderPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Sender)
}

func (t *Transaction) ReceiverPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Receiver)
}
