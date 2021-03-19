package blockchain

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/peerbridge/peerbridge/pkg/encryption"
)

// A block proof of stake.
type Proof struct {
	Challenge            encryption.SHA256
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
		return errors.New(fmt.Sprintf(
			"Hit (%d) is above the upper bound (%d) (Stake: %d)!",
			comparableHit, &proof.UpperBound, proof.Stake,
		))
	}
	// The hit is below the upper bound
	return nil
}
