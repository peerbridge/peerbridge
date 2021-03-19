package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/peerbridge/peerbridge/pkg/encryption"
)

var (
	// The initial block target in the network.
	GenesisTarget uint64 = 100_000
	// The initial block difficulty in the network.
	GenesisDifficulty uint64 = 0

	// The initial block creation time in the network.
	GenesisTimeUnixNano int64 = time.Unix(0, 0).UnixNano()

	// The initial account balances in the network.
	// These account balances are persisted in the genesis block.
	GenesisStake = map[encryption.Secp256k1PublicKey]uint64{}

	// The initial transactions in the genesis block.
	GenesisTransactions = []Transaction{}

	// The initial challenge is a zero byte array.
	GenesisChallenge encryption.SHA256 = encryption.SHA256{}

	// The genesis block.
	GenesisBlock *Block
)

func initGenesisStake() {
	stakeholdersByHexString := map[string]uint64{}
	// Alice
	stakeholdersByHexString["042f135f822ebe8f4af3cb7b47853de0f12251dc46fe22f63f4eb570bd0a7bae1fd744c99079b25b15881ffc7d0f81b206150a6f21e4b8df70acb15df5571c0d47"] = 100_000

	// Bob
	stakeholdersByHexString["041778edef561181eb3192d67154f77ee424817ea8fc0383a715322f9db1539b0083949073eb8fcf698c14b20fca1cda4d0921fdd7a82f0a52116aed7a13d94177"] = 100_000

	for publicKeyHex, stake := range stakeholdersByHexString {
		var publicKey encryption.Secp256k1PublicKey
		bytes, err := hex.DecodeString(publicKeyHex)
		if err != nil {
			panic(err)
		}
		if len(bytes) != encryption.Secp256k1PublicKeyByteLength {
			panic("Invalid secp256k1 public key byte length!")
		}
		var fixedBytes [encryption.Secp256k1PublicKeyByteLength]byte
		copy(fixedBytes[:], bytes[:encryption.Secp256k1PublicKeyByteLength])
		publicKey.Bytes = fixedBytes

		GenesisStake[publicKey] = stake
	}
}

func initGenesisTransactions() {
	for publicKey, stake := range GenesisStake {
		// Generate the genesis transaction ids in a consistent way so that
		// every node has the same starting point
		hasher := sha256.New()
		hasher.Write(publicKey.Bytes[:])
		var id encryption.SHA256
		copy(id.Bytes[:], hasher.Sum(nil)[:encryption.SHA256ByteLength])

		t := Transaction{
			ID:           id,
			Sender:       encryption.Secp256k1PublicKey{}, // Zero address
			Receiver:     publicKey,
			Balance:      stake,
			TimeUnixNano: time.Unix(0, 0).UnixNano(),
			Data:         nil,
			Fee:          0,
		}

		GenesisTransactions = append(GenesisTransactions, t)
	}
}

func initGenesisBlock() {
	GenesisBlock = &Block{
		ID:                   encryption.SHA256{},
		ParentID:             nil,
		TimeUnixNano:         time.Unix(0, 0).UnixNano(),
		Transactions:         GenesisTransactions,
		Creator:              encryption.Secp256k1PublicKey{}, // Zero address
		Target:               &GenesisTarget,
		Challenge:            &GenesisChallenge,
		CumulativeDifficulty: &GenesisDifficulty,
	}
}

func init() {
	initGenesisStake()
	initGenesisTransactions()
	initGenesisBlock()
}
