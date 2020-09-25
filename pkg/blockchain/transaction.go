package blockchain

import (
	"crypto/rand"
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
	Nonce     []byte       `json:"nonce"`
}

func Nonce() (nonce []byte) {
	nonce = make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		panic(err.Error())
	}
	return
}

func (t Transaction) SenderPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Sender)
}

func (t Transaction) ReceiverPublicKey() (*rsa.PublicKey, error) {
	return encryption.PEMStringToPublicKey(t.Receiver)
}
