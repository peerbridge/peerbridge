package blockchain

import (
	"errors"
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

// The request format for the `postTransaction` method.
type PostTransactionRequest struct {
	Transaction *Transaction `json:"transaction"`
}

// The response format for the `postTransaction` method.
type PostTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
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
	var request PostTransactionRequest

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

	err = Instance.AddPendingTransaction(request.Transaction)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	Json(w, r, http.StatusOK, PostTransactionResponse{*request.Transaction})
}

// The response format for the `getTransaction` method.
type GetTransactionResponse struct {
	// The requested transaction.
	Transaction Transaction `json:"transaction"`
}

// Get a transaction (together with its status)
// within the blockchain via http.
//
// This http route returns:
// - 400 BadRequest if the request was malformed
// - 404 NotFound if the transaction could not be found
// - 202 Accepted if the transaction is pending
// - 200 OK together with the transaction and the block
// which includes this transaction
func getTransaction(w http.ResponseWriter, r *http.Request) {
	idParams, ok := r.URL.Query()["id"]

	if !ok || len(idParams[0]) < 1 {
		BadRequest(w, errors.New("The id parameter must be supplied!"))
		return
	}

	requestIDHexString := idParams[0]

	if Instance.ContainsPendingTransactionByID(requestIDHexString) {
		Accepted(w, errors.New("The transaction is still pending!"))
		return
	}

	t, err := Instance.GetTransactionByID(requestIDHexString)
	if err != nil {
		NotFound(w, errors.New("The transaction could not be found!"))
		return
	}

	Json(w, r, http.StatusOK, GetTransactionResponse{*t})
}

// Get an url to the currently active peer.
// This method can be used by other peers to bind to this
// peer via the given multi addresses.
func getPeerURLs(w http.ResponseWriter, r *http.Request) {
	var urls []string
	for _, url := range Peer.URLs {
		urls = append(urls, url.String())
	}
	Json(w, r, http.StatusOK, urls)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/transaction/create", createTransaction)
	router.Get("/transaction/get", getTransaction)

	router.Get("/p2p/urls", getPeerURLs)
	return
}
