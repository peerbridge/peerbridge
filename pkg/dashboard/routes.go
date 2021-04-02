package dashboard

import (
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
	BaseTemplate    = "./templates/dashboard/base.html"
	IndexTemplate   = "./templates/dashboard/index.html"
	BlockTemplate   = "./templates/dashboard/block.html"
	AccountTemplate = "./templates/dashboard/account.html"
)

var templateFunctions = template.FuncMap{
	// General functions
	"shortHex": func(hex string) string {
		return hex[:6]
	},
	"unixToTime": func(unixNano int64) time.Time {
		return time.Unix(0, unixNano)
	},
	// Block functions
	"blockNumberOfTransactions": func(b blockchain.Block) int {
		return len(b.Transactions)
	},
	"blockTimeDiffMillis": func(b1 blockchain.Block, b2 blockchain.Block) int64 {
		return (b2.TimeUnixNano - b1.TimeUnixNano) / 1_000_000
	},
}

var indexViewTemplate = template.Must(template.
	New("base.html").         // This is needed
	Funcs(templateFunctions). // This must be given before ParseFiles
	ParseFiles(BaseTemplate, IndexTemplate),
)

func indexView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	blockchain.Instance.ThreadSafe(func() {
		lastBlockNodes := blockchain.Instance.Head.GetLongestBranch()
		lastBlocks := []blockchain.Block{}
		if len(lastBlockNodes) >= 4 {
			for _, n := range lastBlockNodes[len(lastBlockNodes)-4:] {
				lastBlocks = append(lastBlocks, n.Block)
			}
		} else {
			for _, n := range lastBlockNodes {
				lastBlocks = append(lastBlocks, n.Block)
			}
		}

		publicKey := blockchain.Instance.PublicKey()

		data := struct {
			LastBlocks []blockchain.Block
			PublicKey  string
		}{lastBlocks, publicKey}

		err := indexViewTemplate.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
	})
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

	blockchain.Instance.ThreadSafe(func() {
		block, err := blockchain.Instance.GetBlockByID(requestIDHexString)
		if err != nil {
			NotFound(w, errors.New("The id parameter must be supplied!"))
			return
		}

		// Get parent and map it to the block view model
		var parent *blockchain.Block
		if block.ParentID != nil {
			parent, _ = blockchain.Instance.GetBlockByID(*block.ParentID)
		}

		// Get children and map them to the block view model
		children, _ := blockchain.Instance.GetBlockChildren(requestIDHexString)

		data := struct {
			Block    blockchain.Block
			Parent   *blockchain.Block
			Children *[]blockchain.Block
		}{*block, parent, children}

		err = blockViewTemplate.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
	})
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

	blockchain.Instance.ThreadSafe(func() {
		accountBalance, err := blockchain.Instance.CalculateAccountBalance(requestAccountHexString)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		transactionInfo, err := blockchain.Instance.GetTransactionInfo(requestAccountHexString)
		if err != nil {
			InternalServerError(w, err)
			return
		}

		data := struct {
			PublicKey       secp256k1.PublicKeyHexString
			AccountBalance  int64
			TransactionInfo blockchain.AccountTransactionInfo
		}{requestAccountHexString, *accountBalance, *transactionInfo}

		err = accountViewTemplate.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
	})
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/block", blockView)
	router.Get("/account", accountView)
	router.Get("/ws", BindNewClient)
	router.Get("/", indexView)
	return
}
