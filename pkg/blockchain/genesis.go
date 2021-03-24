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

	// The initial challenge is a zero byte array.
	GenesisChallenge encryption.SHA256 = encryption.SHA256{}

	// The genesis block.
	GenesisBlock *Block
)

func initGenesisKeyPair() {
	keyPair, err := secp256k1.LoadKeyPair("./genesis.key")
	if err != nil {
		panic("Genesis key pair under ./genesis.key missing!")
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
		var publicKey secp256k1.PublicKey
		bytes, err := hex.DecodeString(publicKeyHex)
		if err != nil {
			panic(err)
		}
		if len(bytes) != secp256k1.PublicKeyByteLength {
			panic("Invalid secp256k1 public key byte length!")
		}
		var fixedBytes [secp256k1.PublicKeyByteLength]byte
		copy(fixedBytes[:], bytes[:secp256k1.PublicKeyByteLength])
		publicKey.CompressedBytes = fixedBytes

		// Generate the genesis transaction ids in a consistent way so that
		// every node has the same starting point
		hasher := sha256.New()
		hasher.Write(publicKey.CompressedBytes[:])
		var id encryption.SHA256
		copy(id.Bytes[:], hasher.Sum(nil)[:encryption.SHA256ByteLength])

		t := Transaction{
			ID:           &id,
			Sender:       &GenesisKeyPair.PublicKey,
			Receiver:     &publicKey,
			Balance:      stake,
			TimeUnixNano: time.Unix(0, 0).UnixNano(),
			Data:         nil,
			Fee:          0,
			// Part of the signing process
			Signature: nil,
		}

		signature, err := t.ComputeSignature(&GenesisKeyPair.PrivateKey)
		if err != nil {
			panic(err)
		}
		t.Signature = signature

		GenesisTransactions = append(GenesisTransactions, t)

		log.Printf("Genesis transaction: %X -> Grant %d to %X\n", t.ID.Short(), stake, t.Receiver.Short())
	}
}

func initGenesisBlock() {
	g := Block{
		ID:                   &encryption.SHA256{},
		ParentID:             nil,
		TimeUnixNano:         time.Unix(0, 0).UnixNano(),
		Transactions:         GenesisTransactions,
		Creator:              &GenesisKeyPair.PublicKey,
		Target:               &GenesisTarget,
		Challenge:            &GenesisChallenge,
		CumulativeDifficulty: &GenesisDifficulty,
		// Part of the signature calculation
		Signature: nil,
	}
	signature, err := g.ComputeSignature(&GenesisKeyPair.PrivateKey)
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
