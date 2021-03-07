package blockchain

import (
	"crypto/rsa"
	"time"

	"github.com/peerbridge/peerbridge/pkg/encryption"
)

// The PEM public key format, used as an identification of users.
type PEMPublicKey = string

// A transaction in the blockchain.
// Transactions are obtained via the http interfaces and
// forged into blocks to persist them in the blockchain.
// A transaction has a sender and a receiver, identified
// by their public keys.
type Transaction struct {
	Index      string       `json:"index" pg:"type:uuid,default:gen_random_uuid(),pk,unique,notnull"` // random uuid primary key
	Sender     PEMPublicKey `json:"sender" pg:",notnull"`
	Receiver   PEMPublicKey `json:"receiver" pg:",notnull"`
	Timestamp  time.Time    `json:"timestamp" pg:"default:now(),notnull"`
	Data       []byte       `json:"data" pg:",notnull"` // combined enrypted message of nonce|message|tag as returned by SealedBox.combined
	BlockIndex string       `json:"blockIndex" pg:"type:uuid"`
}

// Generate a RSA public key of a transaction's sender
// from a PEM string.
func (t *Transaction) SenderPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Sender)
}

// Generate a RSA public key of a transaction's receiver
// from a PEM string.
func (t *Transaction) ReceiverPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Receiver)
}
