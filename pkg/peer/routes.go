package peer

import (
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

// Get the peer url for internode p2p connectivity.
func getPeerURLs(w http.ResponseWriter, r *http.Request) {
	Json(w, r, http.StatusOK, PeerURLs)
}

// All specified http routes for the blockchain package.
// Note that calling this method will create a new router.
func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/urls", getPeerURLs)
	return
}
