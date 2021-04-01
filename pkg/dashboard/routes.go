package dashboard

import (
	"net/http"
	"text/template"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	. "github.com/peerbridge/peerbridge/pkg/http"
)

type BlockViewModel struct {
	Block blockchain.Block
}

func (bvm *BlockViewModel) ShortIDString() string {
	return bvm.Block.ID[:6]
}

func (bvm *BlockViewModel) ShortCreatorString() string {
	return bvm.Block.Creator[:6]
}

func (bvm *BlockViewModel) NumberOfTransactions() int {
	return len(bvm.Block.Transactions)
}

func dashboardView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.ParseFiles("./templates/dashboard.html"))

	lastBlockNodes := blockchain.Instance.Head.GetLongestBranch()
	lastBlocks := []BlockViewModel{}
	if len(lastBlockNodes) >= 4 {
		for _, n := range lastBlockNodes[len(lastBlockNodes)-4:] {
			lastBlocks = append(lastBlocks, BlockViewModel{n.Block})
		}
	} else {
		for _, n := range lastBlockNodes {
			lastBlocks = append(lastBlocks, BlockViewModel{n.Block})
		}
	}

	publicKey := blockchain.Instance.PublicKey()

	data := struct {
		LastBlocks []BlockViewModel
		PublicKey  string
	}{lastBlocks, publicKey}

	t.Execute(w, data)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/", dashboardView)
	router.Get("/ws", BindNewClient)
	return
}
