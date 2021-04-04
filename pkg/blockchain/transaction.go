package blockchain

import (
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

// A transaction in the blockchain.
// Transactions are obtained via the http interfaces and
// forged into blocks to persist them in the blockchain.
type Transaction struct {
	// The random id of this transaction, as a unique key.
	ID encryption.SHA256HexString `json:"id" sign:"yes" pg:",pk,unique,notnull"`

	// The sender of this transaction, by address.
	Sender secp256k1.PublicKeyHexString `json:"sender" sign:"yes" pg:",notnull"`

	// The receiver of this transaction, by address.
	Receiver secp256k1.PublicKeyHexString `json:"receiver" sign:"yes" pg:",notnull"`

	// The transferred account balance from the sender
	// to the receiver.
	Balance uint64 `json:"balance" sign:"yes" pg:",notnull,use_zero"`

	// The timestamp of the transaction creation.
	// For the genesis transactions, this is the
	// start of Unix time.
	TimeUnixNano int64 `json:"timeUnixNano" sign:"yes" pg:",notnull,use_zero"`

	// The included transaction data.
	Data *[]byte `json:"data,omitempty" sign:"yes"`

	// The transaction fee.
	Fee uint64 `json:"fee" sign:"yes" pg:",notnull,use_zero"`

	// The signature of the transaction.
	Signature *secp256k1.SignatureHexString `json:"signature" sign:"no" pg:",notnull"`

	// The block id of the block where this transaction is included.
	// This field is `nil` until the transaction is included into
	// a block.
	BlockID *encryption.SHA256HexString `json:"blockID,omitempty" sign:"no" pg:",pk,notnull"`
}
