package blockchain

import (
	"errors"
	"math/big"

	"github.com/peerbridge/peerbridge/pkg/encryption"
)

var (
	ErrProofHitAboveUpperBound = errors.New("Hit is above the upper bound!")
)

// A block proof of stake.
type Proof struct {
	Challenge            encryption.SHA256HexString
	Hit                  uint64
	UpperBound           big.Int
	Target               uint64
	CumulativeDifficulty uint64
	Stake                int64
	NanoSeconds          int64
}

// Check if a calculated proof is valid.
func (proof *Proof) Validate() error {
	// Check if the hit is under the upper bound
	comparableHit := new(big.Int).SetUint64(proof.Hit)
	if comparableHit.Cmp(&proof.UpperBound) == 1 {
		// The hit is above the upper bound
		return ErrProofHitAboveUpperBound
	}
	// The hit is below the upper bound
	return nil
}
