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
	"sync"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
)

const (
	SHA256ByteLength  = 32
	BlockIDByteLength = 16

	// The maximum branch length of the head tree.
	// If a branch exceeds this size, its root will be
	// taken as final and persisted into the blockchain.
	// Note: the bigger the head, the less the probability
	// for a node desync. In production, use 1000 or more
	BlockchainHeadLength = 64
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

	Head *BlockNode

	Tail *[]Block

	lock *sync.Mutex

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
	rootNode := &BlockNode{
		Block:    genesisBlock,
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

// Get a block using its id.
func (chain *Blockchain) GetBlockById(id BlockID) (*Block, error) {
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
func (chain *Blockchain) ContainsBlockByID(id BlockID) bool {
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
	err = chain.ValidateProof(proof)
	if err != nil {
		return nil, err
	}
	// TODO: Check other parameters and the signature
	return proof, nil
}

// Check if the blockchain has a pending block with the given id.
func (chain *Blockchain) ContainsPendingBlockByID(id BlockID) bool {
	for _, pendingBlock := range *chain.PendingBlocks {
		if pendingBlock.ID == id {
			return true
		}
	}
	return false
}

func (chain *Blockchain) AddPendingBlock(b *Block) {
	// Request this block's parent if it is not already
	// in the pending blocks
	if !chain.ContainsPendingBlockByID(b.ParentID) {
		log.Printf("Requested parent block for: %X\n", b.ID[:2])
		Peer.BroadcastNeedsParent(b)
	}

	if chain.ContainsPendingBlockByID(b.ID) {
		return
	}

	*chain.PendingBlocks = append(*chain.PendingBlocks, *b)

	log.Printf(
		"New %s Block %s (H %s, %s T)\n",
		color.Sprintf(fmt.Sprintf("pending"), color.Warning),
		color.Sprintf(fmt.Sprintf("%x", b.ID), color.Debug),
		color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
		color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
	)
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
	if !chain.Head.ContainsBlockByID(b.ParentID) {
		chain.AddPendingBlock(b)
		chain.lock.Unlock()
		return
	}

	// If the block can be added, validate the block
	proof, err := chain.ValidateBlock(b)
	if err != nil {
		log.Printf(
			"Rejected %s Block %s (H %s, %s T) -> Reason: %s\n",
			color.Sprintf("invalid", color.Error),
			color.Sprintf(fmt.Sprintf("%X", b.ID[:2]), color.Debug),
			color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
			color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
			color.Sprintf(fmt.Sprintf("%s", err), color.Error),
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
	chain.Head, chopResult, err = chain.Head.Chop(32)
	if err != nil {
		panic(err)
	}

	// Persist the stem blocks that were chopped off
	for _, n := range *chopResult.StemNodes {
		*chain.Tail = append(*chain.Tail, *n.Block)
	}

	// TODO: recirculate orphaned block transactions

	log.Printf(
		"New %s Block %s -> %s staking %s (H %s, %s T) -> Tail: %s\n",
		color.Sprintf("valid", color.Success),
		color.Sprintf(fmt.Sprintf("%X", b.ParentID[:2]), color.Info),
		color.Sprintf(fmt.Sprintf("%X", b.ID[:2]), color.Debug),
		color.Sprintf(fmt.Sprintf("%d", proof.Stake), color.Success),
		color.Sprintf(fmt.Sprintf("%d", b.Height), color.Info),
		color.Sprintf(fmt.Sprintf("%d", len(b.Transactions)), color.Info),
		color.Sprintf(fmt.Sprintf("%d", len(*chain.Tail)), color.Notice),
	)

	Peer.BroadcastNewBlock(b)

	chain.lock.Unlock()

	// Clean up the pending blocks (i.e. remove all blocks
	// that have a too small height)
	chain.CleanupPendingBlocks()

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
	Stake                int64
}

func (block *Block) AccountBalance(p PublicKey) int64 {
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

// Get the account balance of a public key until a given block.
func (chain *Blockchain) AccountBalanceUntilBlock(
	p PublicKey, id BlockID,
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
	previousBlockNode, err := chain.Head.GetBlockNodeByBlockID(b.ParentID)
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
	accountBalance, err := chain.AccountBalanceUntilBlock(b.Creator, b.ParentID)
	if err != nil {
		return nil, err
	}

	// Note: we use big integers to avoid possible overflows
	// when the upper bound gets very high (e.g. when
	// a node stakes millions in account balance)
	ms := new(big.Int).SetInt64(
		b.Timestamp.Sub(previousBlock.Timestamp).Milliseconds(),
	)
	Tp := new(big.Int).SetUint64(*previousBlock.Target)
	B := new(big.Int).SetInt64(*accountBalance)

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
		Stake:                *accountBalance,
	}, nil
}

// Check if a calculated proof is valid.
func (chain *Blockchain) ValidateProof(proof *Proof) error {
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

// Mint a new block. This returns a block, if the proof of stake
// is successful, otherwise this will return `nil` and an error.
func (chain *Blockchain) MintBlock() (*Block, error) {
	ownPubKey := encryption.PublicKeyToPEMString(&chain.key.PublicKey)
	parentBlock := chain.Head.FindLongestChainEndpoint().Block
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
