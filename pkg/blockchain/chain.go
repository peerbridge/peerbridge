package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

const (
	MaxTransactionsPerBlock = 512
)

var (
	ErrTransactionAlreadyPending = errors.New("Transaction is already pending!")
	ErrTransactionNotFound       = errors.New("Transaction not found!")
	ErrBlockNotFound             = errors.New("Block not found!")
	ErrChildrenNotFound          = errors.New("Children not found!")
	ErrParentBlockNotFound       = errors.New("Parent block not found!")
	ErrAccountHasNoStake         = errors.New("Account has no stake!")
)

type Blockchain struct {
	// The currently pending transactions that were
	// sent to the node (by clients or other nodes)
	// and not yet included in the blockchain.
	PendingTransactions *[]Transaction

	// The blocks that were received or generated
	// but are not yet integrated into the head.
	PendingBlocks *[]Block

	// The account key pair to access the blockchain.
	// This key pair is used to sign blocks and transactions.
	keyPair *secp256k1.KeyPair

	// A lock to ensure mutual exclusion on critical
	// operations that cannot be done concurrently
	// in a safe manner.
	lock sync.Mutex
}

// The main instance of the blockchain.
// This instance is `nil` until `Init(keyPair)` is called.
var Instance *Blockchain

// Initiate a new blockchain with the genesis block.
// The blockchain is accessible under `Instance`.
func InitChain(keyPair *secp256k1.KeyPair) {
	Instance = &Blockchain{
		PendingTransactions: &[]Transaction{},
		PendingBlocks:       &[]Block{},
		keyPair:             keyPair,
	}
}

func (chain *Blockchain) PublicKey() secp256k1.PublicKeyHexString {
	return chain.keyPair.PublicKey
}

func (chain *Blockchain) ThreadSafe(execution func()) {
	chain.lock.Lock()
	defer chain.lock.Unlock()
	execution()
}

// Get a recommended transaction fee based on the pending transactions.
func (chain *Blockchain) RecommendedTransactionFee() int {
	fees := []int{}
	for _, pt := range *chain.PendingTransactions {
		fees = append(fees, int(pt.Fee))
	}
	sort.Ints(fees)
	if len(fees) == 0 {
		return 1
	}
	if len(fees) < MaxTransactionsPerBlock {
		return fees[0] // Min(fees)
	}
	return fees[len(fees)-MaxTransactionsPerBlock] // Min(included)
}

func (chain *Blockchain) Sync(remote string) {
	// TODO: Make this more robust
	// - Handle remote failures, maybe use round robin
	// - Handle chain switches (by fetching all endpoints)

	log.Println("Starting sync...")

	if remote == "" {
		log.Println("No remote provided - sync finished!")
		return
	}

	for {
		foundMoreChildren := false
		endpoint, err := Repo.GetMainChainEndpoint()
		if err != nil {
			panic(err)
		}
		url := fmt.Sprintf("%s/blockchain/blocks/children/get?id=%s", remote, endpoint.ID)

		body := bytes.NewBuffer([]byte{})
		request, err := http.NewRequest("GET", url, body)
		if err != nil {
			panic(err)
		}
		request.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		response, err := client.Do(request)

		if err != nil {
			log.Println(color.Sprintf("Waiting for the remote to be reachable until continuing the sync process...", color.Warning))
			time.Sleep(1 * time.Second)
			continue
		}

		if response.StatusCode == 200 {
			responseBody, err := ioutil.ReadAll(response.Body)
			if err != nil {
				panic(err)
			}

			var response GetChildrenResponse
			err = json.Unmarshal(responseBody, &response)
			if err != nil {
				panic(err)
			}

			for _, child := range *response.Children {
				foundMoreChildren = true
				chain.MigrateBlock(&child, true)
			}
		}

		response.Body.Close()

		if !foundMoreChildren {
			break
		}
	}

	log.Println("Sync finished!")
}

func (chain *Blockchain) ValidateTransaction(t *Transaction) error {
	err := secp256k1.VerifySignature(t, *t.Signature)
	if err != nil {
		return err
	}
	// TODO: Check other parameters
	return nil
}

// Add a given transaction to the pending transactions.
func (chain *Blockchain) AddPendingTransaction(t *Transaction) error {
	if chain.ContainsPendingTransactionByID(t.ID) {
		return ErrTransactionAlreadyPending
	}

	err := chain.ValidateTransaction(t)
	if err != nil {
		return err
	}

	*chain.PendingTransactions = append(*chain.PendingTransactions, *t)

	go BroadcastNewTransaction(t)
	return nil
}

func (chain *Blockchain) RemovePendingTransaction(t *Transaction) {
	if !chain.ContainsPendingTransactionByID(t.ID) {
		return
	}

	newPendingTransactions := []Transaction{}
	for _, pendingTransaction := range *chain.PendingTransactions {
		if pendingTransaction.ID != t.ID {
			newPendingTransactions = append(newPendingTransactions, pendingTransaction)
		}
	}

	*chain.PendingTransactions = newPendingTransactions
}

// Get a pending transaction of the blockchain.
func (chain *Blockchain) GetPendingTransactionByID(id encryption.SHA256HexString) (*Transaction, error) {
	for _, pt := range *chain.PendingTransactions {
		if pt.ID == id {
			return &pt, nil
		}
	}
	return nil, ErrTransactionNotFound
}

func (chain *Blockchain) ContainsPendingTransactionByID(id encryption.SHA256HexString) bool {
	_, err := chain.GetPendingTransactionByID(id)
	return err == nil
}

type AccountTransactionInfo struct {
	PendingTransactions   []Transaction
	PersistedTransactions []Transaction
}

func (chain *Blockchain) GetTransactionInfo(account secp256k1.PublicKeyHexString) (*AccountTransactionInfo, error) {
	accountPendingTxns := []Transaction{}
	for _, t := range *chain.PendingTransactions {
		if t.Sender == account || t.Receiver == account {
			accountPendingTxns = append(accountPendingTxns, t)
		}
	}
	persistedTransactions, err := Repo.GetMainChainTransactionsForAccount(account)
	if err != nil {
		return nil, err
	}
	return &AccountTransactionInfo{
		PendingTransactions:   accountPendingTxns,
		PersistedTransactions: *persistedTransactions,
	}, nil
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
	err = secp256k1.VerifySignature(b, *b.Signature)
	if err != nil {
		return nil, err
	}
	if len(b.Transactions) == 0 {
		return nil, errors.New("No transactions in block!")
	}
	for _, t := range b.Transactions {
		err = chain.ValidateTransaction(&t)
		if err != nil {
			return nil, err
		}
	}
	// TODO: Check other parameters
	return proof, nil
}

// Check if the blockchain has a pending block with the given id.
func (chain *Blockchain) ContainsPendingBlockByID(id encryption.SHA256HexString) bool {
	for _, pendingBlock := range *chain.PendingBlocks {
		if pendingBlock.ID == id {
			return true
		}
	}
	return false
}

func (chain *Blockchain) MigrateBlock(b *Block, syncmode bool) {
	// If the block is already pending, do nothing
	if chain.ContainsPendingBlockByID(b.ID) {
		return
	}

	// Insert the block into the pending blocks
	*chain.PendingBlocks = append([]Block{*b}, *chain.PendingBlocks...)

	// Try to insert pending blocks until none is insertable anymore
	for {
		requeuedBlocks := []Block{}
		invalidBlocks := []Block{}
		insertedBlocks := []Block{}

		for _, pendingB := range *chain.PendingBlocks {
			// Skip blocks that are already in the chain
			if Repo.ContainsBlockByID(pendingB.ID) {
				log.Printf("Dropped block %s (reason: block already in chain, probably rebroadcasted)\n", pendingB.ID[:6])
				invalidBlocks = append(invalidBlocks, pendingB)
				continue
			}

			// Re-queue blocks that need their parent
			if !Repo.ContainsBlockByID(*pendingB.ParentID) {
				log.Printf("Requeued block %s (reason: needs parent)\n", pendingB.ID[:6])
				requeuedBlocks = append(requeuedBlocks, pendingB)
				continue
			}

			// Throw away invalid blocks
			proof, err := chain.ValidateBlock(&pendingB)
			if err != nil {
				log.Printf("Dropped block %s (reason: %s)\n", pendingB.ID[:6], err)
				invalidBlocks = append(invalidBlocks, pendingB)
				continue
			}

			// TODO: Check for duplicated transactions

			err = Repo.AddBlockIfNotExists(&pendingB)
			if err != nil {
				log.Printf("Dropped block %s (reason: insertion not possible)\n", pendingB.ID[:6])
				invalidBlocks = append(invalidBlocks, pendingB)
				continue
			}

			log.Printf(
				"New %s Block %s (Parent: %s) by %s took %sms staking %s (H %s, %s T, S: %s)\n",
				color.Sprintf("valid", color.Success),
				color.Sprintf(fmt.Sprintf("%s", pendingB.ID[:6]), color.Debug),
				color.Sprintf(fmt.Sprintf("%s", (*pendingB.ParentID)[:6]), color.Debug),
				color.Sprintf(fmt.Sprintf("%s", pendingB.Creator[:6]), color.Debug),
				color.Sprintf(fmt.Sprintf("%d", proof.NanoSeconds/1_000_000), color.Debug),
				color.Sprintf(fmt.Sprintf("%d", proof.Stake), color.Success),
				color.Sprintf(fmt.Sprintf("%d", pendingB.Height), color.Info),
				color.Sprintf(fmt.Sprintf("%d", len(pendingB.Transactions)), color.Info),
				color.Sprintf(fmt.Sprintf("%s", (*pendingB.Signature)[:6]), color.Success),
			)

			if !syncmode {
				go BroadcastNewBlock(&pendingB)
			}
			insertedBlocks = append(insertedBlocks, pendingB)

			// Remove pending transactions that were included
			for _, t := range pendingB.Transactions {
				chain.RemovePendingTransaction(&t)
			}
		}

		// Drop all blocks that are pending and
		// transitive children of invalid blocks
		for {
			var indexToDrop *int
		search:
			for i, requeuedBlock := range requeuedBlocks {
				for _, invalidBlock := range invalidBlocks {
					if invalidBlock.ID == *requeuedBlock.ParentID {
						indexToDrop = &i
						break search
					}
				}
			}
			if indexToDrop == nil {
				break
			}
			// Drop transitive child
			requeuedBlocks = append(
				requeuedBlocks[:*indexToDrop],
				requeuedBlocks[*indexToDrop+1:]...,
			)
		}

		chain.PendingBlocks = &requeuedBlocks

		// If we didn't insert any more blocks, stop trying
		if len(insertedBlocks) == 0 {
			break
		}
	}

	// Request all parents that are currently unknown
	if !syncmode {
		for _, block := range *chain.PendingBlocks {
			if !chain.ContainsPendingBlockByID(*block.ParentID) {
				go BroadcastResolveBlockRequest(block.ParentID)
			}
		}
	}
}

func (chain *Blockchain) CalculateProof(b *Block) (*Proof, error) {
	previousBlock, err := Repo.GetBlockByID(*b.ParentID)
	if err != nil {
		return nil, ErrParentBlockNotFound
	}

	challengeHasher := sha256.New()
	hexCreatorBytes, err := hex.DecodeString(b.Creator)
	if err != nil {
		return nil, err
	}
	previousChallengeBytes, err := hex.DecodeString(previousBlock.Challenge)
	if err != nil {
		return nil, err
	}
	challengeHasher.Write(hexCreatorBytes)
	challengeHasher.Write(previousChallengeBytes)
	var challengeBytes [encryption.SHA256ByteLength]byte
	copy(
		challengeBytes[:],
		challengeHasher.Sum(nil)[:encryption.SHA256ByteLength],
	)
	// The hit is used to check if this node is eligible to
	// create a new block (this can be verified by every other node)
	hit := binary.BigEndian.Uint64(challengeBytes[0:8])

	// Get the creator's account balance until the parent block
	// FIXME: Implement a stake height to disallow shuffling attacks
	stake, err := Repo.StakeUntilBlockWithID(b.Creator, previousBlock.ID)
	if err != nil {
		return nil, err
	}
	if *stake <= 0 {
		return nil, ErrAccountHasNoStake
	}

	// Note: we use big integers to avoid possible overflows
	// when the upper bound gets very high (e.g. when
	// a node stakes millions in account balance)
	ns := new(big.Int).SetInt64(
		b.TimeUnixNano - previousBlock.TimeUnixNano,
	)
	Tp := new(big.Int).SetUint64(previousBlock.Target)
	B := new(big.Int).SetInt64(*stake)

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
	Dp := new(big.Int).SetUint64(previousBlock.CumulativeDifficulty)
	CD = CD.Div(new(big.Int).SetUint64(pot), Tn) // pot / Tn
	// TODO: Prevent possible overflows
	CD = CD.Add(Dp, CD) // Dp + (pot / Tn)

	return &Proof{
		Challenge:            hex.EncodeToString(challengeBytes[:]),
		Hit:                  hit,
		UpperBound:           *UB,
		Target:               Tn.Uint64(),
		CumulativeDifficulty: CD.Uint64(),
		Stake:                *stake,
		NanoSeconds:          ns.Int64(),
	}, nil
}

// Get max. 512 transactions from the pending transactions
// which should be minted into a new block, on top of the
// given endpoint block.
func (chain *Blockchain) GetTransactionsToMint(endpointBlock Block) (*[]Transaction, error) {
	blockTransactions := []Transaction{}
	for _, pendingTransaction := range *chain.PendingTransactions {
		if len(blockTransactions) >= MaxTransactionsPerBlock {
			break
		}
		if Repo.ContainsMainChainTransactionByID(pendingTransaction.ID) {
			continue
		}
		blockTransactions = append(blockTransactions, pendingTransaction)
	}
	return &blockTransactions, nil
}

// Mint a new block. This returns a block, if the proof of stake
// is successful, otherwise this will return `nil` and an error.
func (chain *Blockchain) MintBlock() (*Block, error) {
	// Find the longest endpoint block (in the head tree)
	endpointBlock, err := Repo.GetMainChainEndpoint()
	if err != nil {
		return nil, err
	}

	randomID, err := encryption.RandomSHA256HexString()
	if err != nil {
		return nil, err
	}

	// TODO: implement pending transaction ordering (by fee)
	// TODO: implement dynamic block sizes (by transaction amount)
	blockTransactions, err := chain.GetTransactionsToMint(*endpointBlock)
	if err != nil {
		return nil, err
	}
	if len(*blockTransactions) == 0 {
		return nil, errors.New("No transactions to mint!")
	}

	block := &Block{
		ID:           *randomID,
		ParentID:     &endpointBlock.ID,
		Height:       endpointBlock.Height + 1,
		TimeUnixNano: time.Now().UnixNano(),
		Transactions: *blockTransactions,
		Creator:      chain.keyPair.PublicKey,
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
	block.Target = proof.Target
	block.Challenge = proof.Challenge
	block.CumulativeDifficulty = proof.CumulativeDifficulty

	// Signature calculation
	s, err := secp256k1.ComputeSignature(block, chain.keyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	block.Signature = s

	return block, nil
}

// Run a scheduled concurrent block creation loop.
func (chain *Blockchain) RunContinuousMinting() {
	for {
		// Check every 500 ms if we are ready to create a block
		// i.e. if the upper bound is high enough
		time.Sleep(500 * time.Millisecond)
		chain.ThreadSafe(func() {
			block, err := chain.MintBlock()
			if err != nil {
				return
			}
			chain.MigrateBlock(block, false)
		})
	}
}
