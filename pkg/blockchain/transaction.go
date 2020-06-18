package blockchain

import (
	"crypto/rsa"
	"time"

	"github.com/peerbridge/peerbridge/pkg/encryption"
)

type PEMPublicKey = string

type Transaction struct {
	Sender    PEMPublicKey `json:"sender"`
	Receiver  PEMPublicKey `json:"receiver"`
	Timestamp time.Time    `json:"timestamp"`
	Data      []byte       `json:"data"`
}

func (t Transaction) SenderPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Sender)
}

func (t Transaction) ReceiverPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Receiver)
}
