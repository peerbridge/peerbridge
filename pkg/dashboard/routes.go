package dashboard

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	. "github.com/peerbridge/peerbridge/pkg/http"
)

var (
	BaseTemplate        = "./templates/dashboard/base.html"
	IndexTemplate       = "./templates/dashboard/index.html"
	BlockTemplate       = "./templates/dashboard/block.html"
	AccountTemplate     = "./templates/dashboard/account.html"
	TransactionTemplate = "./templates/dashboard/transaction.html"
)

var templateFunctions = template.FuncMap{
	// General functions
	"jsonify": func(v interface{}) string {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return ""
		}
		return string(b)
	},
	"shortHex": func(hex string) string {
		return hex[:6]
	},
	"unixToTime": func(unixNano int64) time.Time {
		return time.Unix(0, unixNano)
	},
	"numBlocksToReward": func(num int) int {
		return num * 100
	},
	// Block functions
	"blockNumberOfTransactions": func(b blockchain.Block) int {
		return len(b.Transactions)
	},
	"blockTimeDiffMillis": func(b1 blockchain.Block, b2 blockchain.Block) int64 {
		return (b2.TimeUnixNano - b1.TimeUnixNano) / 1_000_000
	},
}

func indexView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var indexViewTemplate = template.Must(template.
		New("base.html").         // This is needed
		Funcs(templateFunctions). // This must be given before ParseFiles
		ParseFiles(BaseTemplate, IndexTemplate),
	)

	lastBlocks, err := blockchain.Repo.GetMaxNLastBlocks(12)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	lastTxns, err := blockchain.Repo.GetMaxNLastMainChainTransactions(10)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	publicKey := blockchain.Instance.PublicKey()

	data := struct {
		LastBlocks       []blockchain.Block
		LastTransactions []blockchain.Transaction
		PublicKey        string
	}{*lastBlocks, *lastTxns, publicKey}

	err = indexViewTemplate.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}

var blockViewTemplate = template.Must(template.
	New("base.html").         // This is needed
	Funcs(templateFunctions). // This must be given before ParseFiles
	ParseFiles(BaseTemplate, BlockTemplate),
)

func blockView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	idParams, ok := r.URL.Query()["id"]

	if !ok || len(idParams[0]) < 1 {
		NotFound(w, errors.New("The id parameter must be supplied!"))
		return
	}

	requestIDHexString := idParams[0]

	block, err := blockchain.Repo.GetBlockByID(requestIDHexString)
	if err != nil {
		NotFound(w, errors.New("The id parameter must be supplied!"))
		return
	}

	// Get parent and map it to the block view model
	var parent *blockchain.Block
	if block.ParentID != nil {
		parent, _ = blockchain.Repo.GetBlockByID(*block.ParentID)
	}

	// Get children and map them to the block view model
	children, _ := blockchain.Repo.GetBlockChildren(requestIDHexString)

	data := struct {
		Block    blockchain.Block
		Parent   *blockchain.Block
		Children *[]blockchain.Block
	}{*block, parent, children}

	err = blockViewTemplate.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}

var accountViewTemplate = template.Must(template.
	New("base.html").         // This is needed
	Funcs(templateFunctions). // This must be given before ParseFiles
	ParseFiles(BaseTemplate, AccountTemplate),
)

func accountView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	idParams, ok := r.URL.Query()["id"]

	if !ok || len(idParams[0]) < 1 {
		NotFound(w, errors.New("The id parameter must be supplied!"))
		return
	}

	requestAccountHexString := idParams[0]

	lastBlock, err := blockchain.Repo.GetMainChainEndpoint()
	if err != nil {
		InternalServerError(w, err)
		return
	}

	accountBalance, err := blockchain.Repo.StakeUntilBlockWithID(requestAccountHexString, lastBlock.ID)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	lastForgedBlocks, err := blockchain.Repo.GetMaxNLastBlocksByCreator(12, requestAccountHexString)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	totalForgedBlocks, err := blockchain.Repo.GetBlockCountByCreator(requestAccountHexString)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	blockchain.Instance.ThreadSafe(func() {
		transactionInfo, err := blockchain.Instance.GetTransactionInfo(requestAccountHexString)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		data := struct {
			PublicKey       secp256k1.PublicKeyHexString
			AccountBalance  int64
			TransactionInfo blockchain.AccountTransactionInfo
			LastBlocks      []blockchain.Block
			TotalBlocks     int
		}{requestAccountHexString, *accountBalance, *transactionInfo, *lastForgedBlocks, *totalForgedBlocks}

		err = accountViewTemplate.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
	})
}

var transactionViewTemplate = template.Must(template.
	New("base.html").         // This is needed
	Funcs(templateFunctions). // This must be given before ParseFiles
	ParseFiles(BaseTemplate, TransactionTemplate),
)

func transactionView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	idParams, ok := r.URL.Query()["id"]

	if !ok || len(idParams[0]) < 1 {
		NotFound(w, errors.New("The id parameter must be supplied!"))
		return
	}

	requestTxnID := idParams[0]

	t, err := blockchain.Repo.GetMainChainTransactionByID(requestTxnID)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	b, err := blockchain.Repo.GetBlockByID(*t.BlockID)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	data := struct {
		Transaction blockchain.Transaction
		Block       blockchain.Block
	}{*t, *b}

	err = transactionViewTemplate.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/block", blockView)
	router.Get("/account", accountView)
	router.Get("/transaction", transactionView)
	router.Get("/ws", BindNewClient)
	router.Get("/", indexView)
	return
}
