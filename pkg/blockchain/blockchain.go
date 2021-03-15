package blockchain

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/eventbus"
)

const (
	// The byte length of the SHA256 hashing algorithm.
	SHA256ByteLength = 32
)

type SHA256 = [SHA256ByteLength]byte

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

	// TODO: Add transaction signatures
}

// A block as the main constituent of the blockchain.
type Block struct {
	// The index of the block.
	// The genesis block has index 0. Every following block
	// has the index `ParentIndex + 1`.
	Index uint64 `json:"index"`

	// The index of the parent block.
	// The genesis block has the parent index 0
	// (it is its own ancestor). Otherwise this is
	// `Index -1`.
	ParentIndex uint64 `json:"parentIndex"`

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

	// TODO: Add block signatures
}

type Blockchain struct {
	// The currently pending transactions that were
	// sent to the node (by clients or other nodes)
	// and not yet included in the blockchain.
	PendingTransactions []Transaction
	// The currently forged blocks of the blockchain.
	Blocks []Block

	// The account key to access the blockchain.
	key *rsa.PrivateKey
}

// Create a new blockchain with the genesis block.
func CreateNewBlockchain(key *rsa.PrivateKey) *Blockchain {
	// For now, we act as if the genesis block grants
	// the accessing account some crypto currency.
	// TODO: Replace this with actual private keys of
	// stakeholders
	publicKeyString := encryption.PublicKeyToPEMString(&key.PublicKey)
	genesisTransaction := &Transaction{
		Nonce:     0,
		Sender:    "",
		Receiver:  publicKeyString,
		Balance:   100_000,
		Timestamp: time.Now(),
		Data:      nil,
	}
	// Set the initial target to the maximum uint64
	// to let the blockchain converge to a good value.
	var genesisTarget uint64
	genesisTarget = math.MaxUint64
	genesisBlock := &Block{
		Index:        0,
		ParentIndex:  0,
		Timestamp:    time.Now(),
		Transactions: []Transaction{*genesisTransaction},
		Creator:      "",
		Target:       &genesisTarget,
		// The initial challenge is a zero byte array.
		Challenge: &SHA256{},
	}
	chain := &Blockchain{
		PendingTransactions: []Transaction{},
		Blocks:              []Block{*genesisBlock},
		key:                 key,
	}
	return chain
}

func (chain *Blockchain) ListenOnRemoteUpdates() {
	newRemoteTransactionChannel := eventbus.Instance.
		Subscribe(NewRemoteTransactionTopic)
	newRemoteBlockChannel := eventbus.Instance.
		Subscribe(NewRemoteBlockTopic)

	for {
		select {
		case event := <-newRemoteTransactionChannel:
			if t, castSucceeded := event.Value.(Transaction); castSucceeded {
				chain.AddPendingTransaction(&t)
			}
		case event := <-newRemoteBlockChannel:
			if b, castSucceeded := event.Value.(Block); castSucceeded {
				chain.AddBlock(&b)
			}
		}
	}
}

// Check if the blockchain contains a pending transaction.
func (chain *Blockchain) ContainsPendingTransaction(t *Transaction) bool {
	for _, pt := range chain.PendingTransactions {
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
	chain.PendingTransactions = append(chain.PendingTransactions, *t)

	eventbus.Instance.Publish(NewLocalTransactionTopic, *t)
}

// Get the last block in the blockchain.
func (chain *Blockchain) GetLastBlock() *Block {
	chainLength := len(chain.Blocks)
	if chainLength < 1 {
		panic("The chain should always contain at least the genesis block!")
	}
	return &chain.Blocks[chainLength-1]
}

// Get a block by a given index.
func (chain *Blockchain) GetBlock(index uint64) (*Block, error) {
	for _, cb := range chain.Blocks {
		if index == cb.Index {
			return &cb, nil
		}
	}
	return nil, errors.New("Block not found.")
}

// Get a block by its parent index.
func (chain *Blockchain) GetBlockByParent(index uint64) (*Block, error) {
	for _, cb := range chain.Blocks {
		if index == cb.ParentIndex {
			return &cb, nil
		}
	}
	return nil, errors.New("Block not found.")
}

// Check if the blockchain contains a given block.
func (chain *Blockchain) ContainsBlock(b *Block) bool {
	for _, cb := range chain.Blocks {
		if b.Index == cb.Index {
			return true
		}
	}
	return false
}

// Add a block into the blockchain.
func (chain *Blockchain) AddBlock(b *Block) {
	if chain.ContainsBlock(b) {
		return
	}
	// TODO: Validate block
	chain.Blocks = append(chain.Blocks, *b)

	log.Printf(
		"Created a new block %s.\n",
		color.Sprintf(fmt.Sprintf("%d", b.Index), color.Success),
	)

	eventbus.Instance.Publish(NewLocalBlockTopic, *b)
}

// Get the account balance of a given public key.
func (chain *Blockchain) GetBalance(p *PublicKey) uint64 {
	var accountBalance uint64
	for _, b := range chain.Blocks {
		for _, t := range b.Transactions {
			if t.Sender == *p {
				accountBalance -= t.Balance
			}
			if t.Receiver == *p {
				accountBalance += t.Balance
			}
		}
	}
	return accountBalance
}

// Get the account balance of a given public key until a
// given block index. Note: This block index is excluded from
// the aggregation!
func (chain *Blockchain) GetBalanceUntilBlockIdx(
	i uint64, p *PublicKey,
) uint64 {
	var accountBalance uint64
	for _, bi := range chain.Blocks {
		if bi.Index == i {
			break
		}
		for _, t := range bi.Transactions {
			if t.Sender == *p {
				accountBalance -= t.Balance
			}
			if t.Receiver == *p {
				accountBalance += t.Balance
			}
		}
	}
	return accountBalance
}

// A block proof of stake.
type Proof struct {
	Challenge  SHA256
	Hit        uint64
	UpperBound big.Int
	Target     uint64
}

func (chain *Blockchain) CalculateProof(b *Block) (*Proof, error) {
	previousBlock, err := chain.GetBlock(b.ParentIndex)
	if err != nil {
		return nil, errors.New("No parent block found.")
	}

	challengeHasher := sha256.New()
	challengeHasher.Write([]byte(b.Creator))
	challengeHasher.Write(previousBlock.Challenge[:])
	var challenge SHA256
	copy(challenge[:], challengeHasher.Sum(nil)[:SHA256ByteLength])
	// The hit is used to check if this node is eligible to
	// create a new block (this can be verified by every other node)
	hit := binary.BigEndian.Uint64(challenge[0:8])

	// Get the account balance up until the block (excluding it)
	accountBalance := chain.GetBalanceUntilBlockIdx(b.Index, &b.Creator)

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
	Tn = Tn.Div(Tn, new(big.Int).SetInt64(1000))
	target := Tn.Uint64()

	return &Proof{
		Challenge:  challenge,
		Hit:        hit,
		UpperBound: *UB,
		Target:     target,
	}, nil
}

// Check if a calculated proof is valid.
func (chain *Blockchain) ValidateProof(proof *Proof) bool {
	// Check if the hit is under the upper bound
	comparableHit := new(big.Int).SetUint64(proof.Hit)
	if comparableHit.Cmp(&proof.UpperBound) == 1 {
		// The hit is above the upper bound
		return false
	}
	// The hit is below the upper bound
	return true
}

// Mint a new block. This returns a block, if the proof of stake
// is successful, otherwise this will return `nil` and an error.
func (chain *Blockchain) MintBlock() (*Block, error) {
	ownPubKey := encryption.PublicKeyToPEMString(&chain.key.PublicKey)
	lastBlock := chain.GetLastBlock()
	block := &Block{
		Index:        lastBlock.Index + 1,
		ParentIndex:  lastBlock.Index,
		Timestamp:    time.Now(),
		Transactions: []Transaction{},
		Creator:      ownPubKey,
		// Part of the proof calculation
		Target:    nil,
		Challenge: nil,
	}
	proof, err := chain.CalculateProof(block)
	if err != nil {
		return nil, errors.New("Proof cannot be calculated.")
	}
	if !chain.ValidateProof(proof) {
		return nil, errors.New("Proof is not valid.")
	}
	block.Target = &proof.Target
	block.Challenge = &proof.Challenge
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
