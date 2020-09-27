package blockchain

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/database"
)

func getLastBlock() (*Block, error) {
	blockCount, err := database.Instance.Model((*Block)(nil)).Count()
	if blockCount == 0 {
		return nil, errors.New("The Blockchain is empty.")
	}

	var block Block
	err = database.Instance.Model(&block).
		Order("timestamp ASC").
		Limit(1).
		Select()

	return &block, err
}

func forgeNewBlock(transactions []Transaction) error {
	parentBlock, err := getLastBlock()

	// index and timestamp columns will generated
	newBlock := &Block{}

	if err == nil {
		newBlock.ParentIndex = parentBlock.Index
	} else if err.Error() != "The Blockchain is empty." {
		return err
	}

	newBlock.Transactions = transactions

	if _, err = database.Instance.Model(newBlock).Insert(); err != nil {
		return err
	}

	for _, transaction := range transactions {
		transaction.BlockIndex = newBlock.Index
		if _, err = database.Instance.Model(&transaction).Set("block_index = ?block_index").Where("index = ?index").Update(); err != nil {
			return err
		}
	}

	return nil
}

func getPendingTransactions() (transactions []Transaction, err error) {
	err = database.Instance.Model(&transactions).
		Where("block_index IS NULL").
		Select()

	return
}

func ScheduleBlockCreation(ticker *time.Ticker) {
	for range ticker.C {
		pendingTransactions, err := getPendingTransactions()
		if err != nil {
			log.Printf("Error: %s", color.Sprintf(err.Error(), color.Error))
			return
		}

		if len(pendingTransactions) == 0 {
			continue
		}

		err = forgeNewBlock(pendingTransactions)
		if err != nil {
			log.Printf("Error: %s", color.Sprintf(err.Error(), color.Error))
			return
		}

		blockCount, err := database.Instance.Model((*Block)(nil)).Count()
		if err != nil {
			log.Printf("Error: %s", color.Sprintf(err.Error(), color.Error))
			return
		}

		transactionCount, err := database.Instance.Model((*Transaction)(nil)).Count()
		if err != nil {
			log.Printf("Error: %s", color.Sprintf(err.Error(), color.Error))
			return
		}

		log.Printf(
			"%s. Blocks: %s, Transactions: %s",
			color.Sprintf("Forged a new Block", color.Info),
			color.Sprintf(fmt.Sprintf("%d", blockCount), color.Success),
			color.Sprintf(fmt.Sprintf("%d", transactionCount), color.Warning),
		)
	}
}
