package blockchain

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
)

// A block repository interface to the database.
type BlockRepo struct {
	DB *pg.DB
}

// A default database url that is used to bind the postgres service.
const defaultDatabaseURL = "postgres://postgres:password@localhost:5432/peerbridge?sslmode=disable"

var (
	ErrEmptyBlockRepo = errors.New("The block repository is empty!")
)

// Get a database url from the process environment variables.
// This method is used as a part of database initialization.
// The database url can be configured by setting the
// environment variable `DATABASE_URL`.
func getDatabaseURL() string {
	port := os.Getenv("DATABASE_URL")
	if port != "" {
		return port
	}

	return defaultDatabaseURL
}

func InitializeBlockRepo() *BlockRepo {
	// Initialize the database models
	models := []interface{}{
		(*Block)(nil),
		(*Transaction)(nil),
	}

	opt, err := pg.ParseURL(getDatabaseURL())
	if err != nil {
		panic(err)
	}

	repo := &BlockRepo{pg.Connect(opt)}

	// Initialize the models
	for _, model := range models {
		err := repo.DB.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			panic(err)
		}
	}

	// Insert the genesis block if there are no blocks yet
	blockCount, err := repo.GetBlockCount()
	if err != nil {
		panic(err)
	}
	if *blockCount == 0 {
		err = repo.AddBlock(GenesisBlock)
		if err != nil {
			panic(err)
		}
		_, err := repo.GetBlockByID(GenesisBlock.ID)
		if err != nil {
			panic(err)
		}
		*blockCount += 1
	}
	log.Println(color.Sprintf(fmt.Sprintf("The database contains %d block(s).", *blockCount), color.Info))

	return repo
}

func (r *BlockRepo) GetBlockCount() (*int, error) {
	blockCount, err := r.DB.Model((*Block)(nil)).Count()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &blockCount, nil
}

func (r *BlockRepo) GetLastBlock() (*Block, error) {
	blockCount, err := r.GetBlockCount()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if *blockCount == 0 {
		return nil, ErrEmptyBlockRepo
	}

	var block Block
	err = r.DB.Model(&block).
		Order("height DESC").
		Limit(1).
		Select()

	return &block, err
}

func (r *BlockRepo) GetAllBlocks() ([]Block, error) {
	blocks := []Block{}
	err := r.DB.Model(&blocks).
		Relation("Transactions").
		Select()
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (r *BlockRepo) GetBlockByID(id encryption.SHA256HexString) (*Block, error) {
	var block Block
	err := r.DB.Model(&block).
		Where("id = ?", id).
		Relation("Transactions").
		Select()
	if err != nil {
		return nil, err
	}
	return &block, err
}

func (r *BlockRepo) GetTransactionByID(id encryption.SHA256HexString) (*Transaction, error) {
	var transaction Transaction
	err := r.DB.Model(&transaction).
		Where("id = ?", id).
		Select()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &transaction, nil
}

func (r *BlockRepo) AddBlock(b *Block) error {
	// TODO: Perform consistency checks before addition
	if _, err := r.DB.Model(b).Insert(); err != nil {
		log.Println(err)
		return err
	}
	for _, transaction := range b.Transactions {
		transaction.BlockID = &b.ID
		_, err := r.DB.Model(&transaction).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}
