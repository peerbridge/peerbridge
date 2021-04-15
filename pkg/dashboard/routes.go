package dashboard

import (
	"encoding/json"
	"errors"
	"html/template"
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

type Context struct {
	BaseContext BaseContext
	ViewContext interface{}
}

type BaseContext struct {
	CurrentRecommendedFee int
}

func getBaseContext(completion func(BaseContext)) {
	blockchain.Instance.ThreadSafe(func() {
		fee := blockchain.Instance.RecommendedTransactionFee()
		ctx := BaseContext{
			CurrentRecommendedFee: fee,
		}
		completion(ctx)
	})
}

type TemplateView struct {
	Template   template.Template
	GetContext func(r *http.Request) (viewContext interface{}, err error)
}

func (v *TemplateView) render(w http.ResponseWriter, r *http.Request) {
	getBaseContext(func(baseContext BaseContext) {
		w.Header().Set("Content-Type", "text/html")
		viewContext, err := v.GetContext(r)
		if err != nil {
			InternalServerError(w, err)
			return
		}
		ctx := Context{baseContext, viewContext}
		err = v.Template.Execute(w, ctx)
		if err != nil {
			InternalServerError(w, err)
			return
		}
	})
}

var indexView = TemplateView{
	Template: *template.Must(template.
		New("base.html").         // This is needed
		Funcs(templateFunctions). // This must be given before ParseFiles
		ParseFiles(BaseTemplate, IndexTemplate),
	),
	GetContext: func(r *http.Request) (viewContext interface{}, err error) {
		lastBlocks, err := blockchain.Repo.GetMaxNLastBlocks(12)
		if err != nil {
			return nil, err
		}

		lastTxns, err := blockchain.Repo.GetMaxNLastMainChainTransactions(10)
		if err != nil {
			return nil, err
		}

		publicKey := blockchain.Instance.PublicKey()

		return struct {
			LastBlocks       []blockchain.Block
			LastTransactions []blockchain.Transaction
			PublicKey        string
		}{*lastBlocks, *lastTxns, publicKey}, nil
	},
}

var blockView = TemplateView{
	Template: *template.Must(template.
		New("base.html").         // This is needed
		Funcs(templateFunctions). // This must be given before ParseFiles
		ParseFiles(BaseTemplate, BlockTemplate),
	),
	GetContext: func(r *http.Request) (viewContext interface{}, err error) {
		idParams, ok := r.URL.Query()["id"]

		if !ok || len(idParams[0]) < 1 {
			return nil, errors.New("The id parameter must be supplied!")
		}

		requestIDHexString := idParams[0]

		block, err := blockchain.Repo.GetBlockByID(requestIDHexString)
		if err != nil {
			return nil, err
		}

		// Get parent and map it to the block view model
		var parent *blockchain.Block
		if block.ParentID != nil {
			parent, _ = blockchain.Repo.GetBlockByID(*block.ParentID)
		}

		// Get children and map them to the block view model
		children, _ := blockchain.Repo.GetBlockChildren(requestIDHexString)

		return struct {
			Block    blockchain.Block
			Parent   *blockchain.Block
			Children *[]blockchain.Block
		}{*block, parent, children}, nil
	},
}

var accountView = TemplateView{
	Template: *template.Must(template.
		New("base.html").         // This is needed
		Funcs(templateFunctions). // This must be given before ParseFiles
		ParseFiles(BaseTemplate, AccountTemplate),
	),
	GetContext: func(r *http.Request) (viewContext interface{}, err error) {
		idParams, ok := r.URL.Query()["id"]

		if !ok || len(idParams[0]) < 1 {
			return nil, errors.New("The id parameter must be supplied!")
		}

		requestAccountHexString := idParams[0]

		lastBlock, err := blockchain.Repo.GetMainChainEndpoint()
		if err != nil {
			return nil, err
		}

		accountBalance, err := blockchain.Repo.StakeUntilBlockWithID(requestAccountHexString, lastBlock.ID)
		if err != nil {
			return nil, err
		}

		lastForgedBlocks, err := blockchain.Repo.GetMaxNLastBlocksByCreator(12, requestAccountHexString)
		if err != nil {
			return nil, err
		}

		totalForgedBlocks, err := blockchain.Repo.GetBlockCountByCreator(requestAccountHexString)
		if err != nil {
			return nil, err
		}

		transactionInfo, err := blockchain.Instance.GetTransactionInfo(requestAccountHexString)
		if err != nil {
			return nil, err
		}

		return struct {
			PublicKey       secp256k1.PublicKeyHexString
			AccountBalance  int64
			TransactionInfo blockchain.AccountTransactionInfo
			LastBlocks      []blockchain.Block
			TotalBlocks     int
		}{requestAccountHexString, *accountBalance, *transactionInfo, *lastForgedBlocks, *totalForgedBlocks}, nil
	},
}

var transactionView = TemplateView{
	Template: *template.Must(template.
		New("base.html").         // This is needed
		Funcs(templateFunctions). // This must be given before ParseFiles
		ParseFiles(BaseTemplate, TransactionTemplate),
	),
	GetContext: func(r *http.Request) (viewContext interface{}, err error) {
		idParams, ok := r.URL.Query()["id"]

		if !ok || len(idParams[0]) < 1 {
			return nil, errors.New("The id parameter must be supplied!")
		}

		requestTxnID := idParams[0]

		t, err := blockchain.Repo.GetMainChainTransactionByID(requestTxnID)
		if err != nil {
			return nil, err
		}

		b, err := blockchain.Repo.GetBlockByID(*t.BlockID)
		if err != nil {
			return nil, err
		}

		return struct {
			Transaction blockchain.Transaction
			Block       blockchain.Block
		}{*t, *b}, nil
	},
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/block", blockView.render)
	router.Get("/account", accountView.render)
	router.Get("/transaction", transactionView.render)
	router.Get("/ws", BindNewClient)
	router.Get("/", indexView.render)
	return
}
