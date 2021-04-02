package blockchain

import (
	"errors"
	"html/template"
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

		finalT, err := Instance.GetTransactionByID(requestIDHexString)
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
		children, err := Instance.GetBlockChildren(requestIDHexString)
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
		accountBalance, err := Instance.CalculateAccountBalance(requestAccountHexString)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		Json(w, r, http.StatusOK, GetAccountBalanceResponse{accountBalance})
	})
}

func debugView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := `
<html>
<body style="font-family: monospace">
<h5>Tail Blocks: </h5>
{{range .TailBlocks}}
<p><strong>{{.Height}}</strong> {{.ID}}</p>
{{end}}
<hr>
<h5>Head Block Nodes: </h5>
{{range .HeadBlockNodes}}
<p><strong>{{.Block.Height}}</strong> {{.Block.ID}}</p>
{{end}}
</body>
</html>
	`

	Instance.ThreadSafe(func() {
		tailBlocks, err := Instance.Tail.GetAllBlocks()
		if err != nil {
			InternalServerError(w, err)
			return
		}

		headBlockNodes := Instance.Head.GetLongestBranch()

		data := struct {
			TailBlocks     []Block
			HeadBlockNodes []*BlockTree
		}{tailBlocks, headBlockNodes}

		t := template.Must(template.New("debug-template").Parse(html))
		t.Execute(w, data)
	})
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/transaction/create", createTransaction)
	router.Get("/transaction/get", getTransaction)
	router.Get("/blocks/children/get", getChildBlocks)
	router.Get("/accounts/balance/get", getAccountBalance)
	router.Get("/debug", debugView)
	return
}
