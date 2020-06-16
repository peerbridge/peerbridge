package messaging

import "github.com/peerbridge/peerbridge/pkg/encryption"

type Message struct {
	Signature           encryption.SignatureData
	EncryptedSessionKey encryption.EncryptedData
	EncryptedMessage    []byte
}
