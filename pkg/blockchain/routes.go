package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-pg/pg/v10/orm"
	"github.com/peerbridge/peerbridge/pkg/database"
	. "github.com/peerbridge/peerbridge/pkg/http"
	"github.com/peerbridge/peerbridge/pkg/peer"
)

// Get a transaction via http with a given index.
// The index parameter is supplied as an url parameter.
func getTransaction(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["index"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'index' is missing")
		Text(w, r, http.StatusBadRequest, "Url Param 'index' is missing")
		return
	}

	// Query()["index"] will return an array of items,
	// we only want the single item.
	index := keys[0]

	var transaction Transaction
	err := database.Instance.Model(&transaction).
		Where("index = ?", index).
		Select()

	if err != nil {
		InternalServerError(w, err)
		return
	}

	Json(w, r, http.StatusOK, transaction)
}

// Create a transaction via http.
// The transaction object is passed in the http request
// body in the JSON format.
func createTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction Transaction

	err := DecodeJSONBody(w, r, &transaction)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	_, err = database.Instance.Model(&transaction).Insert()
	if err != nil {
		InternalServerError(w, err)
		return
	}

	// Broadcast the transaction creation to all peers
	// TODO: Use a publish-subscribe scheme for this
	bytes, err := json.Marshal(transaction)
	if err != nil {
		panic(err)
	}
	message := fmt.Sprintf(string(bytes))
	peer.Broadcast(message)

	Json(w, r, http.StatusOK, transaction)
}

// Filter transactions via http.
// The filter parameters must be supplied in the http request body
// in the JSON format. Use the `publicKey` parameter to filter
// all transactions that were received or sent by a given key.
// Use `timestamp` as an optional parameter to only get
// transactions which occured after this timestamp.
// The `timestamp` must be formatted as specified by ISO8601.
func filterTransactions(w http.ResponseWriter, r *http.Request) {
	requestBody := struct {
		PublicKey string
		Timestamp string
	}{}

	err := DecodeJSONBody(w, r, &requestBody)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	transactions := make([]Transaction, 0)
	query := database.Instance.Model(&transactions).WhereGroup(func(q *orm.Query) (*orm.Query, error) {
		q = q.Where("sender = ?", requestBody.PublicKey).WhereOr("receiver = ?", requestBody.PublicKey)
		return q, nil
	})

	if len(requestBody.Timestamp) > 0 {
		query = query.Where("timestamp >= ?", requestBody.Timestamp)
	}

	if err = query.Select(); err != nil {
		InternalServerError(w, err)
		return
	}

	Json(w, r, http.StatusOK, transactions)
}

// Get all received transactions via http.
// The parameters must be supplied in the http request body
// in the JSON format. Use the `publicKey` parameter to filter
// all transactions that were received by a given key.
// Use `timestamp` as an optional parameter to only get
// transactions which occured after this timestamp.
// The `timestamp` must be formatted as specified by ISO8601.
func receivedTransactions(w http.ResponseWriter, r *http.Request) {
	requestBody := struct {
		PublicKey string
		Timestamp string
	}{}

	err := DecodeJSONBody(w, r, &requestBody)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	transactions := make([]Transaction, 0)
	query := database.Instance.Model(&transactions).Where("receiver = ?", requestBody.PublicKey)

	if len(requestBody.Timestamp) > 0 {
		query = query.Where("timestamp >= ?", requestBody.Timestamp)
	}

	if err = query.Select(); err != nil {
		InternalServerError(w, err)
		return
	}

	Json(w, r, http.StatusOK, transactions)
}

// Get all blocks in the blockchain via http.
func allBlocks(w http.ResponseWriter, r *http.Request) {
	blocks := make([]Block, 0)
	err := database.Instance.Model(&blocks).
		Relation("Transactions"). // Fetch associated Transactions
		Select()

	if err != nil {
		InternalServerError(w, err)
		return
	}

	Json(w, r, http.StatusOK, blocks)
}

// Get a specific block in the blockchain via http.
// The `index` parameter to select the block is given
// as an url parameter.
func getBlock(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["index"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'index' is missing")
		Text(w, r, http.StatusBadRequest, "Url Param 'index' is missing")
		return
	}

	// Query()["index"] will return an array of items,
	// we only want the single item.
	index := keys[0]

	var block Block
	err := database.Instance.Model(&block).
		Where("index = ?", index).
		Relation("Transactions"). // ORM - Fetch associated Transactions
		Select()

	if err != nil {
		InternalServerError(w, err)
		return
	}

	Json(w, r, http.StatusOK, block)
}

// All specified http routes for the blockchain package.
// Note that calling this method will create a new router.
func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/transactions/new", createTransaction)
	router.Post("/transactions/filter", filterTransactions)
	router.Post("/transactions/received", receivedTransactions)
	router.Get("/transactions", getTransaction)
	router.Get("/blocks/all", allBlocks)
	router.Get("/blocks", getBlock)
	return
}
