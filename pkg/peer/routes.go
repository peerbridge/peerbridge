package peer

import (
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

// Get an url to the currently active peer.
func getPeerURLs(w http.ResponseWriter, r *http.Request) {
	var urls []string
	for _, url := range Instance.URLs {
		urls = append(urls, url.String())
	}
	Json(w, r, http.StatusOK, urls)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/urls", getPeerURLs)
	return
}
