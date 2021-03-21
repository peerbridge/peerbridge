package blockchain

import (
	"encoding/json"
	"reflect"

	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

// A block as the main constituent of the blockchain.
type Block struct {
	// The random id of the block.
	ID encryption.SHA256 `json:"id" sign:"yes"`

	// The id of the parent block.
	// This is only "nil" for the genesis block.
	ParentID *encryption.SHA256 `json:"parentID" sign:"yes"`

	// The height of the block.
	// The genesis block has height 0.
	Height uint64 `json:"height" sign:"yes"`

	// The timestamp of the block creation.
	// For the genesis block, this is the
	// start of Unix time.
	TimeUnixNano int64 `json:"timeUnixNano" sign:"yes"`

	// The transactions that are included in the block.
	// This includes regular transactions from clients
	// and a special reward transaction at the block end.
	Transactions []Transaction `json:"transactions" sign:"yes"`

	// The address of the block creator.
	Creator secp256k1.PublicKey `json:"creator" sign:"yes"`

	// The target value of this block which has to be met
	// by the block creator.
	Target *uint64 `json:"target" sign:"yes"`

	// The challenge is created by signing the parent block challenge
	// with the block creator public keyand hashing it with the
	// SHA256 hashing algorithm. The challenge is used to
	// determine if an account is eligible to create a new block.
	Challenge *encryption.SHA256 `json:"challenge" sign:"yes"`

	// The cumulative difficulty of this block increases
	// over the chain length with regards of the base target.
	// It is used to determine which chain to use when
	// there are two chains with equal maximum heights.
	CumulativeDifficulty *uint64 `json:"cumulativeDifficulty" sign:"yes"`

	// The signature of the block.
	Signature *secp256k1.Signature `json:"signature" sign:"no"`
}

func (block *Block) AccountBalance(p secp256k1.PublicKey) int64 {
	var accountBalance int64 = 0
	if block.Creator.Equals(&p) {
		accountBalance += 100 // Block reward
	}
	for _, t := range block.Transactions {
		if t.Receiver.Equals(&p) {
			// FIXME: Theoretically, this could overflow
			// with very high balances
			accountBalance += int64(t.Balance)
		}
		if t.Sender.Equals(&p) {
			// FIXME: Theoretically, this could overflow
			// with very high balances
			accountBalance -= int64(t.Balance)
			accountBalance -= int64(t.Fee)
		}
	}
	return accountBalance
}

func (block *Block) GetSigningInput() (*secp256k1.SigningInput, error) {
	// Get all fields that are tagged with sign:"yes"
	t := reflect.TypeOf(*block)
	v := reflect.ValueOf(*block)
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

func (block *Block) ComputeSignature(
	p *secp256k1.PrivateKey,
) (*secp256k1.Signature, error) {
	input, err := block.GetSigningInput()
	if err != nil {
		return nil, err
	}
	return input.Sign(p)
}

func (block *Block) VerifySignature() error {
	input, err := block.GetSigningInput()
	if err != nil {
		return err
	}
	return input.VerifySignature(block.Signature, &block.Creator)
}
