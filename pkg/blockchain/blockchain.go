package blockchain

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
)

const (
	SHA256ByteLength  = 32
	BlockIDByteLength = 16
)

type SHA256 = [SHA256ByteLength]byte

type BlockID = [BlockIDByteLength]byte

type PublicKey = string

// A transaction in the blockchain.
// Transactions are obtained via the http interfaces and
// forged into blocks to persist them in the blockchain.
type Transaction struct {
	// The nonce of this transaction, as a unique key.
	Nonce uint64 `json:"nonce"`

	// The sender of this transaction, by address.
	Sender PublicKey `json:"sender"`

	// The receiver of this transaction, by address.
	Receiver PublicKey `json:"receiver"`

	// The transferred account balance from the sender
	// to the receiver.
	Balance uint64 `json:"balance"`

	// The time of creation for this transaction.
	Timestamp time.Time `json:"timestamp"`

	// The included transaction data.
	Data *[]byte `json:"data"`

	// The transaction fee.
	Fee uint64 `json:"fee"`

	// TODO: Add transaction signatures
}

// A block as the main constituent of the blockchain.
type Block struct {
	// The random id of the block.
	ID BlockID `json:"id"`

	// The id of the parent block.
	ParentID BlockID `json:"parentID"`

	// The height of the block.
	// The genesis block has height 0.
	Height uint64 `json:"height"`

	// The timestamp of the block creation.
	// For the genesis block, this is the
	// start of Unix time.
	Timestamp time.Time `json:"timestamp"`

	// The transactions that are included in the block.
	// This includes regular transactions from clients
	// and a special reward transaction at the block end.
	Transactions []Transaction `json:"transactions"`

	// The address of the block creator.
	Creator PublicKey `json:"creator"`

	// The target value of this block which has to be met
	// by the block creator.
	Target *uint64 `json:"target"`

	// The challenge is created by signing the parent block challenge
	// with the block creator public keyand hashing it with the
	// SHA256 hashing algorithm. The challenge is used to
	// determine if an account is eligible to create a new block.
	Challenge *SHA256 `json:"challenge"`

	// The cumulative difficulty of this block increases
	// over the chain length with regards of the base target.
	// It is used to determine which chain to use when
	// there are two chains with equal maximum heights.
	CumulativeDifficulty *uint64 `json:"cumulativeDifficulty"`

	// TODO: Add block signatures
}

type Blockchain struct {
	// The currently pending transactions that were
	// sent to the node (by clients or other nodes)
	// and not yet included in the blockchain.
	PendingTransactions *[]Transaction

	PendingBlocks *[]Block

	RootNode BlockNode

	// The account key to access the blockchain.
	key *rsa.PrivateKey
}

var Instance *Blockchain

// Initiate a new blockchain with the genesis block.
// The blockchain is accessible under `Instance`.
func Init(key *rsa.PrivateKey) {
	// TODO: Replace this with actual private keys of
	// stakeholders
	aliceTransaction := &Transaction{
		Nonce:     0,
		Sender:    "",
		Receiver:  encryption.AliceExamplePublicKey(),
		Balance:   100_000,
		Timestamp: time.Unix(0, 0),
		Fee:       0,
		Data:      nil,
	}
	bobTransaction := &Transaction{
		Nonce:     0,
		Sender:    "",
		Receiver:  encryption.BobExamplePublicKey(),
		Balance:   100_000,
		Timestamp: time.Unix(0, 0),
		Fee:       0,
		Data:      nil,
	}
	var genesisTarget uint64 = 100_000
	var genesisDifficulty uint64 = 0
	genesisBlock := &Block{
		ID:        BlockID{1},
		ParentID:  BlockID{0},
		Timestamp: time.Unix(0, 0),
		Transactions: []Transaction{
			*aliceTransaction,
			*bobTransaction,
		},
		Creator: "",
		Target:  &genesisTarget,
		// The initial challenge is a zero byte array.
		Challenge:            &SHA256{},
		CumulativeDifficulty: &genesisDifficulty,
	}
	rootNode := BlockNode{
		Block:    genesisBlock,
		Children: &[]BlockNode{},
		Parent:   nil,
	}
	Instance = &Blockchain{
		PendingTransactions: &[]Transaction{},
		PendingBlocks:       &[]Block{},
		RootNode:            rootNode,
		key:                 key,
	}
}

// Check if the blockchain contains a pending transaction.
func (chain *Blockchain) ContainsPendingTransaction(t *Transaction) bool {
	for _, pt := range *chain.PendingTransactions {
		if t.Nonce == pt.Nonce {
			return true
		}
	}
	return false
}

// Add a given transaction to the pending transactions.
func (chain *Blockchain) AddPendingTransaction(t *Transaction) {
	if chain.ContainsPendingTransaction(t) {
		return
	}
	// TODO: Validate transaction
	*chain.PendingTransactions = append(*chain.PendingTransactions, *t)

	Peer.BroadcastNewTransaction(t)
}

func (chain *Blockchain) ValidateBlock(b *Block) error {
	proof, err := chain.CalculateProof(b)
	if err != nil {
		return err
	}
	err = chain.ValidateProof(proof)
	if err != nil {
		return err
	}
	// TODO: Check other parameters and the signature

	return nil
}

// Add a block into the blockchain.
func (chain *Blockchain) AddBlock(b *Block) {
	// If the block is already in the blockchain, do nothing.
	if chain.RootNode.ContainsBlock(b) {
		return
	}

	// Check if the block can be added to the blockchain
	if !chain.RootNode.ContainsBlockByID(b.ParentID) {
		for _, pendingBlock := range *chain.PendingBlocks {
			if pendingBlock.ID == b.ID {
				return
			}
		}
		log.Printf(
			"New %s Block %s (H %s, %s T)\n",
			color.Sprintf(fmt.Sprintf("pending"), color.Warning),
			color.Sprintf(fmt.Sprintf("%x", b.ID), color.Debug),
			color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
			color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
		)
		// Add the block to the pending blocks and request its parent
		*chain.PendingBlocks = append(*chain.PendingBlocks, *b)
		Peer.BroadcastNeedsParent(b)
		return
	}

	// If the block was in the pending blocks, remove it
	pendingBlocksWOB := []Block{}
	for _, pendingB := range *chain.PendingBlocks {
		if pendingB.ID != b.ID {
			pendingBlocksWOB = append(pendingBlocksWOB, pendingB)
		}
	}
	*chain.PendingBlocks = pendingBlocksWOB

	// If the block can be added, validate the block
	err := chain.ValidateBlock(b)
	if err != nil {
		log.Printf(
			"New %s Block %s (H %s, %s T) -> %s\n",
			color.Sprintf("invalid", color.Error),
			color.Sprintf(fmt.Sprintf("%x", b.ID), color.Debug),
			color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
			color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
			color.Sprintf(fmt.Sprintf("%s", err), color.Error),
		)
		return
	}

	_, err = chain.RootNode.InsertBlock(b)
	if err != nil {
		// There should be no error since we checked before if the
		// block is actually insertable
		panic(err)
	}

	log.Printf(
		"New %s Block %s (H %s, %s T)\n",
		color.Sprintf("valid", color.Success),
		color.Sprintf(fmt.Sprintf("%x", b.ID), color.Debug),
		color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
		color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
	)

	Peer.BroadcastNewBlock(b)

	// Integrate the pending blocks if possible
	for _, block := range *chain.PendingBlocks {
		chain.AddBlock(&block)
	}
}

// A block proof of stake.
type Proof struct {
	Challenge            SHA256
	Hit                  uint64
	UpperBound           big.Int
	Target               uint64
	CumulativeDifficulty uint64
}

func (chain *Blockchain) CalculateProof(b *Block) (*Proof, error) {
	previousBlockNode, err := chain.RootNode.GetBlockNodeByBlockID(b.ParentID)
	if err != nil {
		return nil, errors.New("No parent block found.")
	}
	previousBlock := previousBlockNode.Block

	challengeHasher := sha256.New()
	challengeHasher.Write([]byte(b.Creator))
	challengeHasher.Write(previousBlock.Challenge[:])
	var challenge SHA256
	copy(challenge[:], challengeHasher.Sum(nil)[:SHA256ByteLength])
	// The hit is used to check if this node is eligible to
	// create a new block (this can be verified by every other node)
	hit := binary.BigEndian.Uint64(challenge[0:8])

	// Get the account balance up until the block (excluding it)
	// TODO: Compute actual account balances
	var accountBalance uint64 = 100_000

	// Note: we use big integers to avoid possible overflows
	// when the upper bound gets very high (e.g. when
	// a node stakes millions in account balance)
	ms := new(big.Int).SetInt64(
		b.Timestamp.Sub(previousBlock.Timestamp).Milliseconds(),
	)
	Tp := new(big.Int).SetUint64(*previousBlock.Target)
	B := new(big.Int).SetUint64(accountBalance)

	// Upper Bound = (Tp * ms * B) / 1000
	UB := new(big.Int)
	UB = UB.Mul(Tp, ms)
	UB = UB.Mul(UB, B)
	UB = UB.Div(UB, new(big.Int).SetInt64(1000))

	// New Block Target = (Tp * ms) / 1000
	Tn := new(big.Int)
	Tn = Tn.Mul(Tp, ms)
	// TODO: Prevent possible overflows
	Tn = Tn.Div(Tn, new(big.Int).SetInt64(1000))

	// New Block Cumulative Difficulty = Dp + (pot / Tn)
	// Where Pot = 2^64
	CD := new(big.Int)
	var pot uint64 = 1 << 63
	Dp := new(big.Int).SetUint64(*previousBlock.CumulativeDifficulty)
	CD = CD.Div(new(big.Int).SetUint64(pot), Tn) // pot / Tn
	// TODO: Prevent possible overflows
	CD = CD.Add(Dp, CD) // Dp + (pot / Tn)

	return &Proof{
		Challenge:            challenge,
		Hit:                  hit,
		UpperBound:           *UB,
		Target:               Tn.Uint64(),
		CumulativeDifficulty: CD.Uint64(),
	}, nil
}

// Check if a calculated proof is valid.
func (chain *Blockchain) ValidateProof(proof *Proof) error {
	// Check if the hit is under the upper bound
	comparableHit := new(big.Int).SetUint64(proof.Hit)
	if comparableHit.Cmp(&proof.UpperBound) == 1 {
		// The hit is above the upper bound
		return errors.New(fmt.Sprintf(
			"Hit (%d) is above the upper bound (%d)!",
			comparableHit, &proof.UpperBound,
		))
	}
	// The hit is below the upper bound
	return nil
}

// Mint a new block. This returns a block, if the proof of stake
// is successful, otherwise this will return `nil` and an error.
func (chain *Blockchain) MintBlock() (*Block, error) {
	ownPubKey := encryption.PublicKeyToPEMString(&chain.key.PublicKey)
	parentBlock := chain.RootNode.FindLongestChainEndpoint().Block
	randomID := BlockID{}
	rand.Seed(time.Now().UTC().UnixNano())
	_, err := rand.Read(randomID[:])
	if err != nil {
		return nil, err
	}
	block := &Block{
		ID:           randomID,
		ParentID:     parentBlock.ID,
		Height:       parentBlock.Height + 1,
		Timestamp:    time.Now(),
		Transactions: []Transaction{},
		Creator:      ownPubKey,
		// Part of the proof calculation
		Target:               nil,
		Challenge:            nil,
		CumulativeDifficulty: nil,
	}
	proof, err := chain.CalculateProof(block)
	if err != nil {
		return nil, err
	}
	err = chain.ValidateProof(proof)
	if err != nil {
		return nil, err
	}
	block.Target = &proof.Target
	block.Challenge = &proof.Challenge
	block.CumulativeDifficulty = &proof.CumulativeDifficulty
	return block, nil
}

// Run a scheduled block creation loop.
func (chain *Blockchain) RunContinuousMinting() {
	for {
		// Check every 500 ms if we are ready to create a block
		// i.e. if the upper bound is high enough
		time.Sleep(500 * time.Millisecond)
		block, err := chain.MintBlock()
		if err != nil {
			continue
		}
		chain.AddBlock(block)
	}
}
