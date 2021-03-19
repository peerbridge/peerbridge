package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

const (
	// The maximum branch length of the head tree.
	// If a branch exceeds this size, its root will be
	// taken as final and persisted into the blockchain.
	// Note: the bigger the head, the less the probability
	// for a node desync. In production, use 1000 or more
	BlockchainHeadLength = 256
)

type Blockchain struct {
	// The currently pending transactions that were
	// sent to the node (by clients or other nodes)
	// and not yet included in the blockchain.
	PendingTransactions *[]Transaction

	PendingBlocks *[]Block

	Head *BlockNode

	Tail *[]Block

	lock *sync.Mutex

	// The account key pair to access the blockchain.
	// This key pair is used to sign blocks and transactions.
	keyPair *secp256k1.KeyPair
}

var Instance *Blockchain

// Initiate a new blockchain with the genesis block.
// The blockchain is accessible under `Instance`.
func Init(keyPair *secp256k1.KeyPair) {
	rootNode := &BlockNode{
		Block:    GenesisBlock,
		Children: &[]*BlockNode{},
		Parent:   nil,
		lock:     &sync.Mutex{},
	}
	Instance = &Blockchain{
		PendingTransactions: &[]Transaction{},
		PendingBlocks:       &[]Block{},
		Head:                rootNode,
		Tail:                &[]Block{},
		lock:                &sync.Mutex{},
		keyPair:             keyPair,
	}
	log.Printf(
		"Initialized blockchain. Our public key: %s...\n",
		color.Sprintf(fmt.Sprintf("%X", keyPair.PublicKey.Short()), color.Notice),
	)
}

// Check if the blockchain contains a pending transaction.
func (chain *Blockchain) ContainsPendingTransaction(t *Transaction) bool {
	for _, pt := range *chain.PendingTransactions {
		if t.ID == pt.ID {
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

// Get a block using its id.
func (chain *Blockchain) GetBlockById(id encryption.SHA256) (*Block, error) {
	// Check the head tree first.
	node, err := chain.Head.GetBlockNodeByBlockID(id)
	if err == nil {
		return node.Block, nil
	}

	// Check the tail list afterwards.
	for _, b := range *chain.Tail {
		if b.ID == id {
			return &b, nil
		}
	}

	return nil, errors.New("Block not found!")
}

// Check if the blockchain contains a given block (by id).
func (chain *Blockchain) ContainsBlockByID(id encryption.SHA256) bool {
	_, err := chain.GetBlockById(id)
	return err == nil
}

// Check if the blockchain contains a given block.
func (chain *Blockchain) ContainsBlock(b *Block) bool {
	return chain.ContainsBlockByID(b.ID)
}

func (chain *Blockchain) ValidateBlock(b *Block) (*Proof, error) {
	proof, err := chain.CalculateProof(b)
	if err != nil {
		return nil, err
	}
	err = proof.Validate()
	if err != nil {
		return nil, err
	}
	err = b.VerifySignature()
	if err != nil {
		return nil, err
	}
	// TODO: Check other parameters
	return proof, nil
}

// Check if the blockchain has a pending block with the given id.
func (chain *Blockchain) ContainsPendingBlockByID(id encryption.SHA256) bool {
	for _, pendingBlock := range *chain.PendingBlocks {
		if pendingBlock.ID == id {
			return true
		}
	}
	return false
}

func (chain *Blockchain) AddPendingBlock(b *Block) {
	if b.ParentID == nil {
		return
	}
	// Request this block's parent if it is not already
	// in the pending blocks
	if !chain.ContainsPendingBlockByID(*b.ParentID) {
		log.Printf("Requested parent block for: %X\n", b.ID.Short())
		Peer.BroadcastNeedsParent(b)
	}

	if chain.ContainsPendingBlockByID(b.ID) {
		return
	}

	*chain.PendingBlocks = append(*chain.PendingBlocks, *b)
}

func (chain *Blockchain) RemovePendingBlock(b *Block) {
	newPendingBlocks := &[]Block{}
	for _, pendingB := range *chain.PendingBlocks {
		if pendingB.ID != b.ID {
			*newPendingBlocks = append(*newPendingBlocks, pendingB)
		}
	}
	chain.PendingBlocks = newPendingBlocks
}

// Get the height of the last block in the blockchain tail.
func (chain *Blockchain) TailHeight() uint64 {
	if chain.Tail == nil {
		return 0
	}
	return uint64(len(*chain.Tail))
}

// Remove all pending blocks that will never be added to the blockchain.
// That is, if their height is too small to be incorporated into the head.
func (chain *Blockchain) CleanupPendingBlocks() {
	tailHeight := chain.TailHeight()
	newPendingBlocks := &[]Block{}
	for _, pendingB := range *chain.PendingBlocks {
		if pendingB.Height > tailHeight {
			*newPendingBlocks = append(*newPendingBlocks, pendingB)
		}
	}
	chain.PendingBlocks = newPendingBlocks
}

// Add a block into the blockchain.
func (chain *Blockchain) AddBlock(b *Block) {
	chain.lock.Lock()

	// If the block is already in the blockchain, do nothing
	if chain.ContainsBlock(b) {
		chain.lock.Unlock()
		return
	}

	// If the block is too far away from the current height, reject it
	if chain.TailHeight() >= b.Height {
		chain.lock.Unlock()
		return
	}

	// Check if the block can be added to the blockchain
	// That is, if the parent block is in the blockchain
	// head tree. Otherwise, add it to the pending blocks
	if !chain.Head.ContainsBlockByID(*b.ParentID) {
		chain.AddPendingBlock(b)
		chain.lock.Unlock()
		return
	}

	// If the block can be added, validate the block
	proof, err := chain.ValidateBlock(b)
	if err != nil {
		log.Printf(
			"New %s Block %s by %s (H %s, %s T) -> Tail: %s, Validation Error: %s\n",
			color.Sprintf("invalid", color.Error),
			color.Sprintf(fmt.Sprintf("%X", b.ID.Short()), color.Debug),
			color.Sprintf(fmt.Sprintf("%X", b.Creator.Short()), color.Debug),
			color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
			color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
			color.Sprintf(fmt.Sprintf("%d", len(*chain.Tail)), color.Notice),
			color.Sprintf(err, color.Error),
		)
		chain.lock.Unlock()
		return
	}

	// Insert the block into the chain head tree (by its parent)
	_, err = chain.Head.InsertBlock(b)
	if err != nil {
		// There should be no error since we checked before if the
		// block is actually insertable
		panic(err)
	}

	// If the block was pending, remove it from the pending blocks
	if chain.ContainsPendingBlockByID(b.ID) {
		chain.RemovePendingBlock(b)
	}

	// Chop the chain head and keep it nice and short
	var chopResult *ChopResult
	chain.Head, chopResult, err = chain.Head.Chop(BlockchainHeadLength)
	if err != nil {
		panic(err)
	}

	// Persist the stem blocks that were chopped off
	for _, n := range *chopResult.StemNodes {
		*chain.Tail = append(*chain.Tail, *n.Block)
	}

	// TODO: recirculate orphaned block transactions

	log.Printf(
		"New %s Block %s by %s took %sms staking %s (H %s, %s T, S: %s) -> Tail: %s\n",
		color.Sprintf("valid", color.Success),
		color.Sprintf(fmt.Sprintf("%X", b.ID.Short()), color.Debug),
		color.Sprintf(fmt.Sprintf("%X", b.Creator.Short()), color.Debug),
		color.Sprintf(fmt.Sprintf("%d", proof.NanoSeconds/1_000_000), color.Debug),
		color.Sprintf(fmt.Sprintf("%d", proof.Stake), color.Success),
		color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
		color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
		color.Sprintf(fmt.Sprintf("%X", b.Signature.Short()), color.Success),
		color.Sprintf(fmt.Sprintf("%d", len(*chain.Tail)), color.Notice),
	)

	Peer.BroadcastNewBlock(b)

	chain.lock.Unlock()

	chain.CleanupPendingBlocks()

	// Integrate the pending blocks if possible
	for _, block := range *chain.PendingBlocks {
		chain.AddBlock(&block)
	}
}

// Get the account balance of a public key until a given block.
func (chain *Blockchain) AccountBalanceUntilBlock(
	p secp256k1.PublicKey, id encryption.SHA256,
) (*int64, error) {
	var accountBalance int64 = 0
	for _, b := range *chain.Tail {
		accountBalance += b.AccountBalance(p)
		if b.ID == id {
			return &accountBalance, nil
		}
	}
	headchain, err := chain.Head.GetChain(id)
	if err != nil {
		return nil, err
	}
	for _, bn := range *headchain {
		accountBalance += bn.Block.AccountBalance(p)
		if bn.Block.ID == id {
			return &accountBalance, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Block %x not found!", id))
}

func (chain *Blockchain) CalculateProof(b *Block) (*Proof, error) {
	previousBlockNode, err := chain.Head.GetBlockNodeByBlockID(*b.ParentID)
	if err != nil {
		return nil, errors.New("No parent block found.")
	}
	previousBlock := previousBlockNode.Block

	challengeHasher := sha256.New()
	challengeHasher.Write(b.Creator.CompressedBytes[:])
	challengeHasher.Write(previousBlock.Challenge.Bytes[:])
	var challenge encryption.SHA256
	copy(
		challenge.Bytes[:],
		challengeHasher.Sum(nil)[:encryption.SHA256ByteLength],
	)
	// The hit is used to check if this node is eligible to
	// create a new block (this can be verified by every other node)
	hit := binary.BigEndian.Uint64(challenge.Bytes[0:8])

	// Get the account balance up until the block (excluding it)
	accountBalance, err := chain.AccountBalanceUntilBlock(
		b.Creator, *b.ParentID,
	)
	if err != nil {
		return nil, err
	}

	if *accountBalance <= 0 {
		return nil, errors.New("Account has no stake!")
	}

	// Note: we use big integers to avoid possible overflows
	// when the upper bound gets very high (e.g. when
	// a node stakes millions in account balance)
	ns := new(big.Int).SetInt64(
		b.TimeUnixNano - previousBlock.TimeUnixNano,
	)
	Tp := new(big.Int).SetUint64(*previousBlock.Target)
	B := new(big.Int).SetInt64(*accountBalance)

	// Upper Bound = (Tp * ns * B) / (1 * 10^9)
	UB := new(big.Int)
	UB = UB.Mul(Tp, ns)
	UB = UB.Mul(UB, B)
	UB = UB.Div(UB, new(big.Int).SetInt64(1_000_000_000))

	// New Block Target = (Tp * ns) / (1 * 10^9)
	Tn := new(big.Int)
	Tn = Tn.Mul(Tp, ns)
	// TODO: Prevent possible overflows
	Tn = Tn.Div(Tn, new(big.Int).SetInt64(1_000_000_000))

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
		Stake:                *accountBalance,
		NanoSeconds:          ns.Int64(),
	}, nil
}

// Mint a new block. This returns a block, if the proof of stake
// is successful, otherwise this will return `nil` and an error.
func (chain *Blockchain) MintBlock() (*Block, error) {
	parentBlock := chain.Head.FindLongestChainEndpoint().Block
	randomID, err := encryption.RandomSHA256()
	if err != nil {
		return nil, err
	}
	block := &Block{
		ID:           *randomID,
		ParentID:     &parentBlock.ID,
		Height:       parentBlock.Height + 1,
		TimeUnixNano: time.Now().UnixNano(),
		Transactions: []Transaction{},
		Creator:      chain.keyPair.PublicKey,

		// Part of the proof calculation
		Target:               nil,
		Challenge:            nil,
		CumulativeDifficulty: nil,

		// Part of the signature computation
		Signature: nil,
	}

	// Proof calculation
	proof, err := chain.CalculateProof(block)
	if err != nil {
		return nil, err
	}
	err = proof.Validate()
	if err != nil {
		return nil, err
	}
	block.Target = &proof.Target
	block.Challenge = &proof.Challenge
	block.CumulativeDifficulty = &proof.CumulativeDifficulty

	// Signature calculation
	s, err := block.ComputeSignature(&chain.keyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	block.Signature = s

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
