package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
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

// The blockchain repo instance.
var Repo *BlockRepo

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

func InitRepo() {
	// Initialize the database models
	models := []interface{}{
		(*Block)(nil),
		(*Transaction)(nil),
	}

	dbURL := getDatabaseURL()
	opt, err := pg.ParseURL(dbURL)
	if err != nil {
		panic(err)
	}

	log.Println(color.Sprintf(fmt.Sprintf("Connecting to database under: %s", dbURL), color.Notice))

	repo := BlockRepo{DB: pg.Connect(opt)}

	// Poll until the database is alive
	ctx := context.Background()
	for {
		err := repo.DB.Ping(ctx)
		if err == nil {
			break
		}
		log.Println(color.Sprintf("Waiting until the database is online...", color.Warning))
		time.Sleep(time.Second * 1)
	}

	// Initialize the models
	for _, model := range models {
		err := repo.DB.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			panic(err)
		}
	}

	err = repo.AddBlockIfNotExists(GenesisBlock)
	if err != nil {
		panic(err)
	}
	blockCount, err := repo.GetBlockCount()
	log.Println(color.Sprintf(fmt.Sprintf("The database contains %d block(s).", *blockCount), color.Info))

	Repo = &repo
}

func (r *BlockRepo) GetBlockCount() (*int, error) {
	blockCount, err := r.DB.Model((*Block)(nil)).Count()
	if err != nil {
		return nil, err
	}
	return &blockCount, nil
}

func (r *BlockRepo) GetBlockCountByCreator(creator secp256k1.PublicKeyHexString) (*int, error) {
	blockCount, err := r.DB.Model((*Block)(nil)).
		Where("creator = ?", creator).
		Count()
	if err != nil {
		return nil, err
	}
	return &blockCount, nil
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

func (r *BlockRepo) ContainsBlockByID(id encryption.SHA256HexString) bool {
	_, err := r.GetBlockByID(id)
	return err == nil
}

func (r *BlockRepo) GetBlockChildren(id encryption.SHA256HexString) (*[]Block, error) {
	blocks := []Block{}
	err := r.DB.Model(&blocks).
		Where("parent_id = ?", id).
		Relation("Transactions").
		Select()
	if err != nil {
		return nil, err
	}
	return &blocks, err
}

func (r *BlockRepo) GetMaxNLastBlocks(n int) (*[]Block, error) {
	var blocks []Block
	err := r.DB.Model(&blocks).
		Order("height DESC", "cumulative_difficulty DESC").
		Limit(n).
		Relation("Transactions").
		Select()
	if err != nil {
		return nil, err
	}
	return &blocks, nil
}

func (r *BlockRepo) GetMaxNLastBlocksByCreator(n int, creator secp256k1.PublicKeyHexString) (*[]Block, error) {
	var blocks []Block
	err := r.DB.Model(&blocks).
		Order("height DESC", "cumulative_difficulty DESC").
		Limit(n).
		Relation("Transactions").
		Where("creator = ?", creator).
		Select()
	if err != nil {
		return nil, err
	}
	return &blocks, nil
}

func (r *BlockRepo) GetMainChainEndpoint() (*Block, error) {
	var block Block
	err := r.DB.Model(&block).
		Order("height DESC", "cumulative_difficulty DESC").
		Limit(1).
		Relation("Transactions").
		Select()
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *BlockRepo) GetChainToBlock(b Block) (*[]Block, error) {
	var blocks []Block
	_, err := r.DB.Query(&blocks, `
		WITH RECURSIVE chain AS(
			SELECT *
			FROM blocks
			WHERE id = ?
			UNION ALL
			SELECT b.*
			FROM blocks b
			INNER JOIN chain c
			ON c.parent_id = b.id
		)

		SELECT id
		FROM chain
		ORDER BY height ASC;
	`, b.ID)

	// Populate all fields
	r.DB.Model(&blocks).WherePK().Relation("Transactions").Select()
	if err != nil {
		return nil, err
	}
	return &blocks, nil
}

// A partial sql query to fetch the main chain.
// This is a recursive query, which first fetches
// the chain endpoint (by highest height and cumulative
// difficulty) and then reconstructs the parent path.
var mainChainPartialQuery = `
	WITH RECURSIVE endpoint AS(
		SELECT *
		FROM blocks
		ORDER BY height DESC, cumulative_difficulty DESC
		LIMIT 1
	), main_chain AS(
		SELECT *
		FROM endpoint
		UNION ALL
		SELECT b.*
		FROM blocks b
		INNER JOIN main_chain c
		ON c.parent_id = b.id
	)
`

func (r *BlockRepo) GetMaxNLastMainChainTransactions(n int) (*[]Transaction, error) {
	var txns []Transaction
	_, err := r.DB.Query(&txns, fmt.Sprintf(`
		%s

		SELECT t.*
		FROM transactions t
		INNER JOIN main_chain c ON t.block_id = c.id
		ORDER BY time_unix_nano DESC
		LIMIT ?;
	`, mainChainPartialQuery), n)
	if err != nil {
		return nil, err
	}
	return &txns, nil
}

func (r *BlockRepo) GetMainChainTransactionByID(id encryption.SHA256HexString) (*Transaction, error) {
	var txns []Transaction
	_, err := r.DB.Query(&txns, fmt.Sprintf(`
		%s

		SELECT t.*
		FROM transactions t
		INNER JOIN main_chain c ON t.block_id = c.id
		WHERE t.id = ?;
	`, mainChainPartialQuery), id)
	if err != nil {
		return nil, err
	}
	if len(txns) > 1 {
		return nil, errors.New("Multiple transactions returned!")
	}
	if len(txns) == 0 {
		return nil, errors.New("Transaction not found!")
	}
	return &txns[0], nil
}

func (r *BlockRepo) ContainsMainChainTransactionByID(id encryption.SHA256HexString) bool {
	_, err := r.GetMainChainTransactionByID(id)
	return err == nil
}

func (r *BlockRepo) GetMainChainTransactionsForAccount(account secp256k1.PublicKeyHexString) (*[]Transaction, error) {
	var txns []Transaction
	_, err := r.DB.Query(&txns, fmt.Sprintf(`
		%s

		SELECT t.*
		FROM transactions t
		INNER JOIN main_chain c ON t.block_id = c.id
		WHERE ? IN (t.sender, t.receiver)
		ORDER BY time_unix_nano DESC;
	`, mainChainPartialQuery), account)
	if err != nil {
		return nil, err
	}
	return &txns, nil
}

func (r *BlockRepo) AddBlockIfNotExists(b *Block) error {
	// TODO: Perform consistency checks before addition
	if _, err := r.DB.Model(b).OnConflict("DO NOTHING").Insert(); err != nil {
		return err
	}
	for _, transaction := range b.Transactions {
		transaction.BlockID = &b.ID
		_, err := r.DB.Model(&transaction).OnConflict("DO NOTHING").Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

// Compute the stake of an account until a block id.
func (r *BlockRepo) StakeUntilBlockWithID(
	p secp256k1.PublicKeyHexString,
	blockID encryption.SHA256HexString,
) (*int64, error) {
	block, err := r.GetBlockByID(blockID)
	if err != nil {
		return nil, err
	}

	chain, err := r.GetChainToBlock(*block)
	if err != nil {
		return nil, err
	}

	stake := int64(0)

	// TODO: Use ORM to do this computation
	for _, b := range *chain {
		if b.Creator == p {
			stake += 100 // Block reward
		}
		for _, t := range b.Transactions {
			if t.Sender == p {
				// FIXME: Theoretically, this could overflow
				// with very high fees or balances
				stake -= int64(t.Balance)
				stake -= int64(t.Fee)
			}
			if t.Receiver == p {
				// FIXME: Theoretically, this could overflow
				// with very high balances
				stake += int64(t.Balance)
			}
		}
	}

	return &stake, nil
}
