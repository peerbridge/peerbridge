package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/block"
)

func generateBlock(w http.ResponseWriter, r *http.Request) (*Block, bool) {
	var b Block

	err := decodeJSONBody(w, r, &b)

	if err != nil {
		var re *RequestError
		if errors.As(err, &re) {
			http.Error(w, re.msg, re.status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return nil, false
	}

	return &b, true
}

func newBlock(w http.ResponseWriter, r *http.Request) {
	if b, ok := generateBlock(w, r); ok {
		fmt.Fprintf(w, "Created new Block: %+v", b)
	}
}

func hashBlock(w http.ResponseWriter, r *http.Request) {
	if b, ok := generateBlock(w, r); ok {
		fmt.Fprintf(w, "%+s", b.Hash())
	}
}

var Routes = []Route{
	Route{method: http.MethodPost, pattern: "/block/new", handler: http.HandlerFunc(newBlock)},
	Route{method: http.MethodPost, pattern: "/block/hash", handler: http.HandlerFunc(hashBlock)},
}
