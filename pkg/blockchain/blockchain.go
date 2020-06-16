package blockchain

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
)

var MainBlockChain BlockChain

type BlockChain struct {
	Blocks              []Block
	PendingTransactions []Transaction
}

func (c *BlockChain) addBlock(b Block) {
	c.Blocks = append(c.Blocks, b)
}

func (c *BlockChain) getLastBlock() (*Block, error) {
	if len(c.Blocks) == 0 {
		return nil, errors.New("The Blockchain is empty.")
	}
	return &c.Blocks[len(c.Blocks)-1], nil
}

func (c *BlockChain) ForgeNewBlock() *Block {
	parent, err := c.getLastBlock()
	var newBlock Block
	if err == nil {
		newBlock = Block{parent.Index + 1, time.Now(), parent.Hash(), c.PendingTransactions}
	} else {
		genesisHash := sha256.Sum256([]byte("Skrrrt"))
		newBlock = Block{0, time.Now(), genesisHash, c.PendingTransactions}
	}
	c.PendingTransactions = []Transaction{}
	c.addBlock(newBlock)
	return &newBlock
}

func (c *BlockChain) AddTransaction(t Transaction) {
	c.PendingTransactions = append(c.PendingTransactions, t)
}

func (c *BlockChain) GetAllForgedTransactions() (t []Transaction) {
	for _, block := range c.Blocks {
		for _, transaction := range block.Transactions {
			t = append(t, transaction)
		}
	}
	return
}

// Get transactions for a given public key.
func (c *BlockChain) GetForgedTransactions(k string) (t []Transaction) {
	for _, transaction := range c.GetAllForgedTransactions() {
		if transaction.Receiver == k || transaction.Sender == k {
			t = append(t, transaction)
		}
	}
	return
}

func ScheduleBlockCreation(ticker *time.Ticker) {
	for range ticker.C {
		if len(MainBlockChain.PendingTransactions) == 0 {
			continue
		}
		log.Printf(
			"%s. Blocks: %s, Transactions: %s",
			color.Sprintf("Forging a new Block", color.Info),
			color.Sprintf(fmt.Sprintf("%d", len(MainBlockChain.Blocks)), color.Success),
			color.Sprintf(fmt.Sprintf("%d", len(MainBlockChain.GetAllForgedTransactions())), color.Warning),
		)
		MainBlockChain.ForgeNewBlock()
	}
}
