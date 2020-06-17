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

type FilterTransactionsRequest struct {
	PublicKey string
}

func filterTransactions(w http.ResponseWriter, r *http.Request) {
	var requestData FilterTransactionsRequest

	err := DecodeJSONBody(w, r, &requestData)
	if err != nil {
		fmt.Println(err)
		InternalServerError(w, err)
		return
	}

	transactions := MainBlockChain.GetForgedTransactions(requestData.PublicKey)
	Json(w, r, http.StatusCreated, transactions)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/transactions/new", createTransaction)
	router.Post("/transactions/filter", filterTransactions)
	return
}
