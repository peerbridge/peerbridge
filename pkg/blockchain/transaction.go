package blockchain

import (
	"github.com/peerbridge/peerbridge/pkg/encryption"
)

// A transaction in the blockchain.
// Transactions are obtained via the http interfaces and
// forged into blocks to persist them in the blockchain.
type Transaction struct {
	// The random id of this transaction, as a unique key.
	ID encryption.SHA256 `json:"id"`

	// The sender of this transaction, by address.
	Sender encryption.Secp256k1PublicKey `json:"sender"`

	// The receiver of this transaction, by address.
	Receiver encryption.Secp256k1PublicKey `json:"receiver"`

	// The transferred account balance from the sender
	// to the receiver.
	Balance uint64 `json:"balance"`

	// The timestamp of the transaction creation.
	// For the genesis transactions, this is the
	// start of Unix time.
	TimeUnixNano int64 `json:"timeUnixNano"`

	// The included transaction data.
	Data *[]byte `json:"data"`

	// The transaction fee.
	Fee uint64 `json:"fee"`

	// TODO: Add transaction signatures
}
