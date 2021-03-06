package blockchain

import (
	"errors"
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

// The request format for the `postTransaction` method.
type CreateTransactionRequest struct {
	Transaction *Transaction `json:"transaction"`
}

// The response format for the `postTransaction` method.
type CreateTransactionResponse struct {
	Transaction *Transaction `json:"transaction"`
}

// Create a new transaction in the blockchain via http.
// This adds the transaction to the transaction queue.
// To check if the transaction has been included
// into the blockchain, use the method `getTransaction`.
//
// This http route returns:
// - 400 BadRequest if the request was malformed
// - 500 InternalServerError if the transaction could not be added
// - 200 OK if the transaction was added to the queue
func createTransaction(w http.ResponseWriter, r *http.Request) {
	var request CreateTransactionRequest

	err := DecodeJSONBody(w, r, &request)
	if err != nil {
		NotFound(w, err)
		return
	}

	if request.Transaction == nil {
		NotFound(w, errors.New("Transaction could not be decoded!"))
		return
	}

	// TODO: Validate transaction

	Instance.ThreadSafe(func() {
		err := Instance.AddPendingTransaction(request.Transaction)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		Json(w, r, http.StatusOK, CreateTransactionResponse{request.Transaction})
	})
}

// The response format for the `getTransaction` method.
type GetTransactionResponse struct {
	// The requested transaction.
	Transaction *Transaction `json:"transaction"`
}

// Get a transaction (together with its status)
// within the blockchain via http.
//
// This http route returns:
// - 400 BadRequest if the request was malformed
// - 404 NotFound if the transaction could not be found
// - 202 Accepted together with the transaction if the transaction is pending
// - 200 OK together with the transaction
// which includes this transaction
func getTransaction(w http.ResponseWriter, r *http.Request) {
	idParams, ok := r.URL.Query()["id"]

	if !ok || len(idParams[0]) < 1 {
		BadRequest(w, errors.New("The id parameter must be supplied!"))
		return
	}

	requestIDHexString := idParams[0]

	Instance.ThreadSafe(func() {
		pendingT, err := Instance.GetPendingTransactionByID(requestIDHexString)
		if err == nil {
			Json(w, r, http.StatusAccepted, GetTransactionResponse{pendingT})
			return
		}

		finalT, err := Repo.GetMainChainTransactionByID(requestIDHexString)
		if err != nil {
			NotFound(w, errors.New("The transaction could not be found!"))
			return
		}

		Json(w, r, http.StatusOK, GetTransactionResponse{finalT})
	})
}

// The response format for the `getTransaction` method.
type GetChildrenResponse struct {
	Children *[]Block `json:"children"`
}

// Get children of a block within the blockchain via http.
//
// This http route returns:
// - 400 BadRequest if the request was malformed
// - 404 NotFound if the block or its children could not be found
// - 200 OK together with the child blocks
func getChildBlocks(w http.ResponseWriter, r *http.Request) {
	idParams, ok := r.URL.Query()["id"]

	if !ok || len(idParams[0]) < 1 {
		BadRequest(w, errors.New("The id parameter must be supplied!"))
		return
	}

	requestIDHexString := idParams[0]

	Instance.ThreadSafe(func() {
		children, err := Repo.GetBlockChildren(requestIDHexString)
		if err != nil {
			NotFound(w, errors.New("Children not found!"))
			return
		}

		Json(w, r, http.StatusOK, GetChildrenResponse{children})
	})
}

// The response format for the `getAccountBalance` method.
type GetAccountBalanceResponse struct {
	Balance *int64 `json:"balance"`
}

// Get a user's account balance (within the longest chain) via http.
//
// This http route returns:
// - 400 BadRequest if the request was malformed
// - 500 InternalServerError if the balance could not be calculated
// - 200 OK together with the user's balance
func getAccountBalance(w http.ResponseWriter, r *http.Request) {
	accountParams, ok := r.URL.Query()["account"]

	if !ok || len(accountParams[0]) < 1 {
		BadRequest(w, errors.New("The account parameter must be supplied!"))
		return
	}

	requestAccountHexString := accountParams[0]

	Instance.ThreadSafe(func() {
		lastBlock, err := Repo.GetMainChainEndpoint()
		if err != nil {
			InternalServerError(w, err)
			return
		}

		accountBalance, err := Repo.StakeUntilBlockWithID(requestAccountHexString, lastBlock.ID)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		Json(w, r, http.StatusOK, GetAccountBalanceResponse{accountBalance})
	})
}

type GetRecommendedTransactionFeeResponse struct {
	Fee *int `json:"fee"`
}

func getRecommendedTransactionFee(w http.ResponseWriter, r *http.Request) {
	Instance.ThreadSafe(func() {
		fee := Instance.RecommendedTransactionFee()
		response := GetRecommendedTransactionFeeResponse{&fee}
		Json(w, r, http.StatusOK, response)
	})
}

// The response format for the `getAccountTransactions` method.
type GetAccountTransactionsResponse struct {
	// The requested transaction.
	Transactions *[]Transaction `json:"transactions"`
}

func getAccountTransactions(w http.ResponseWriter, r *http.Request) {
	accountParams, ok := r.URL.Query()["account"]

	if !ok || len(accountParams[0]) < 1 {
		BadRequest(w, errors.New("The account parameter must be supplied!"))
		return
	}

	requestAccountHexString := accountParams[0]

	txns, err := Repo.GetMainChainTransactionsForAccount(requestAccountHexString)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	Json(w, r, http.StatusOK, GetAccountTransactionsResponse{txns})
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/transaction/create", createTransaction)
	router.Get("/transaction/get", getTransaction)
	router.Get("/fees/get", getRecommendedTransactionFee)
	router.Get("/blocks/children/get", getChildBlocks)
	router.Get("/accounts/balance/get", getAccountBalance)
	router.Get("/accounts/transactions/get", getAccountTransactions)
	return
}
