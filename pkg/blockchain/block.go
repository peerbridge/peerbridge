package blockchain

import "github.com/peerbridge/peerbridge/pkg/encryption"

// A block as the main constituent of the blockchain.
type Block struct {
	// The random id of the block.
	ID encryption.SHA256 `json:"id"`

	// The id of the parent block.
	// This is only "nil" for the genesis block.
	ParentID *encryption.SHA256 `json:"parentID"`

	// The height of the block.
	// The genesis block has height 0.
	Height uint64 `json:"height"`

	// The timestamp of the block creation.
	// For the genesis block, this is the
	// start of Unix time.
	TimeUnixNano int64 `json:"timeUnixNano"`

	// The transactions that are included in the block.
	// This includes regular transactions from clients
	// and a special reward transaction at the block end.
	Transactions []Transaction `json:"transactions"`

	// The address of the block creator.
	Creator encryption.Secp256k1PublicKey `json:"creator"`

	// The target value of this block which has to be met
	// by the block creator.
	Target *uint64 `json:"target"`

	// The challenge is created by signing the parent block challenge
	// with the block creator public keyand hashing it with the
	// SHA256 hashing algorithm. The challenge is used to
	// determine if an account is eligible to create a new block.
	Challenge *encryption.SHA256 `json:"challenge"`

	// The cumulative difficulty of this block increases
	// over the chain length with regards of the base target.
	// It is used to determine which chain to use when
	// there are two chains with equal maximum heights.
	CumulativeDifficulty *uint64 `json:"cumulativeDifficulty"`

	// TODO: Add block signatures
}

func (block *Block) AccountBalance(p encryption.Secp256k1PublicKey) int64 {
	var accountBalance int64 = 0
	if block.Creator == p {
		accountBalance += 100 // Block reward
	}
	for _, t := range block.Transactions {
		if t.Receiver == p {
			// FIXME: Theoretically, this could overflow
			// with very high balances
			accountBalance += int64(t.Balance)
		}
		if t.Sender == p {
			// FIXME: Theoretically, this could overflow
			// with very high balances
			accountBalance -= int64(t.Balance)
			accountBalance -= int64(t.Fee)
		}
	}
	return accountBalance
}
