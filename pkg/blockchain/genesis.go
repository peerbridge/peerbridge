package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

var (
	// The initial block height.
	GenesisHeight uint64 = 0
	// The initial block target in the network.
	GenesisTarget uint64 = 100_000
	// The initial block difficulty in the network.
	GenesisDifficulty uint64 = 0

	// The initial block creation time in the network.
	GenesisTimeUnixNano int64 = time.Unix(0, 0).UnixNano()

	// The genesis account that created the genesis block.
	GenesisKeyPair *secp256k1.KeyPair

	// The initial transactions in the genesis block.
	GenesisTransactions = []Transaction{}

	GenesisChallenge encryption.SHA256HexString = encryption.ZeroSHA256HexString()

	GenesisAddress encryption.SHA256HexString = encryption.ZeroSHA256HexString()

	// The genesis block.
	GenesisBlock *Block
)

func initGenesisKeyPair() {
	keyPair, err := secp256k1.LoadKeyPair("./genesis.key.json")
	if err != nil {
		panic("Genesis key pair under ./genesis.key.json missing!")
	}
	GenesisKeyPair = keyPair
}

func initGenesisTransactions() {
	stakeholdersByHexString := map[string]uint64{}
	// Alice
	stakeholdersByHexString["0372689db204d56d9bb7122497eef4732cce308b73f3923fc076aed3c2dfa4ad04"] = 100_000

	// Bob
	stakeholdersByHexString["03f1f2fbd80b49b8ffc8194ac0a0e0b7cf0c7e21bca2482c5fba7adf67db41dec5"] = 100_000

	for publicKeyHex, stake := range stakeholdersByHexString {
		// Generate the genesis transaction ids in a consistent way so that
		// every node has the same starting point
		hasher := sha256.New()
		publicKeyBytes, err := hex.DecodeString(publicKeyHex)
		if err != nil {
			panic(err)
		}
		hasher.Write(publicKeyBytes)
		var id [encryption.SHA256ByteLength]byte
		copy(id[:], hasher.Sum(nil)[:encryption.SHA256ByteLength])
		idHex := hex.EncodeToString(id[:])

		t := Transaction{
			ID:           idHex,
			Sender:       GenesisKeyPair.PublicKey,
			Receiver:     publicKeyHex,
			Balance:      stake,
			TimeUnixNano: time.Unix(0, 0).UnixNano(),
			Data:         nil,
			Fee:          0,
			BlockID:      &GenesisAddress, // Genesis block
			// Part of the signing process
			Signature: nil,
		}

		signature, err := secp256k1.ComputeSignature(t, GenesisKeyPair.PrivateKey)
		if err != nil {
			panic(err)
		}
		t.Signature = signature

		GenesisTransactions = append(GenesisTransactions, t)

		log.Printf("Genesis transaction: %s -> Grant %d to %s\n", t.ID[:6], stake, t.Receiver[:6])
	}
}

func initGenesisBlock() {
	g := Block{
		ID:                   GenesisAddress,
		ParentID:             nil,
		Height:               GenesisHeight,
		TimeUnixNano:         time.Unix(0, 0).UnixNano(),
		Transactions:         GenesisTransactions,
		Creator:              GenesisKeyPair.PublicKey,
		Target:               GenesisTarget,
		Challenge:            GenesisChallenge,
		CumulativeDifficulty: GenesisDifficulty,
		// Part of the signature calculation
		Signature: nil,
	}
	signature, err := secp256k1.ComputeSignature(g, GenesisKeyPair.PrivateKey)
	if err != nil {
		panic(err)
	}
	g.Signature = signature
	GenesisBlock = &g
}

func init() {
	initGenesisKeyPair()
	initGenesisTransactions()
	initGenesisBlock()
}
