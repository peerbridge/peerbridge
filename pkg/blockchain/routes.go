package blockchain

import (
	"fmt"
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

func createTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction Transaction

	err := DecodeJSONBody(w, r, &transaction)
	if err != nil {
		fmt.Println(err)
		InternalServerError(w, err)
		return
	}

	MainBlockChain.AddTransaction(transaction)
	Json(w, r, http.StatusCreated, transaction)
}

type PublicKeyRequest struct {
	PublicKey string
}

func filterTransactions(w http.ResponseWriter, r *http.Request) {
	var requestData PublicKeyRequest

	err := DecodeJSONBody(w, r, &requestData)
	if err != nil {
		fmt.Println(err)
		InternalServerError(w, err)
		return
	}

	transactions := MainBlockChain.GetForgedTransactions(requestData.PublicKey)
	Json(w, r, http.StatusCreated, transactions)
}

func receivedTransactions(w http.ResponseWriter, r *http.Request) {
	var requestData PublicKeyRequest

	err := DecodeJSONBody(w, r, &requestData)
	if err != nil {
		fmt.Println(err)
		InternalServerError(w, err)
		return
	}

	transactions := MainBlockChain.GetReceivedTransactions(requestData.PublicKey)
	Json(w, r, http.StatusCreated, transactions)
}

func allBlocks(w http.ResponseWriter, r *http.Request) {
	Json(w, r, http.StatusCreated, MainBlockChain.Blocks)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/transactions/new", createTransaction)
	router.Post("/transactions/filter", filterTransactions)
	router.Post("/transactions/received", receivedTransactions)
	router.Post("/blocks/all", allBlocks)
	return
}
