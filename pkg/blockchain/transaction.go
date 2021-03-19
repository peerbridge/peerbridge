package blockchain

import (
	"encoding/json"
	"reflect"

	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

// A transaction in the blockchain.
// Transactions are obtained via the http interfaces and
// forged into blocks to persist them in the blockchain.
type Transaction struct {
	// The random id of this transaction, as a unique key.
	ID encryption.SHA256 `json:"id"`

	// The sender of this transaction, by address.
	Sender secp256k1.PublicKey `json:"sender"`

	// The receiver of this transaction, by address.
	Receiver secp256k1.PublicKey `json:"receiver"`

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

	// The signature of the transaction.
	Signature *secp256k1.Signature `json:"signature" sign:"no"`
}

func (tx *Transaction) GetSigningInput() (*secp256k1.SigningInput, error) {
	// Get all fields that are tagged with sign:"yes"
	t := reflect.TypeOf(*tx)
	v := reflect.ValueOf(*tx)
	values := []interface{}{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("sign") == "yes" {
			values = append(values, v.Field(i).Interface())
		}
	}
	// Marshal those fields to json and use
	// it to create the signing input
	bytes, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	input := secp256k1.NewSigningInput(bytes)
	return &input, nil
}

func (tx *Transaction) ComputeSignature(
	p *secp256k1.PrivateKey,
) (*secp256k1.Signature, error) {
	input, err := tx.GetSigningInput()
	if err != nil {
		return nil, err
	}
	return input.Sign(p)
}

func (tx *Transaction) VerifySignature() error {
	input, err := tx.GetSigningInput()
	if err != nil {
		return err
	}
	return input.VerifySignature(tx.Signature, &tx.Sender)
}
