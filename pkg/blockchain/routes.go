package blockchain

import (
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

type GetBlockAfterRequest struct {
	ID BlockID `json:"id"`
}

type GetBlockAfterResponse struct {
	Block *Block `json:"block"`
}

func getBlockAfter(w http.ResponseWriter, r *http.Request) {
	var request GetBlockAfterRequest
	err := DecodeJSONBody(w, r, &request)
	if err != nil {
		InternalServerError(w, err)
	}

	var response GetBlockAfterResponse
	block, err := Instance.GetBlockByParent(request.ID)
	if err != nil {
		NotFound(w, err)
	}
	response.Block = block
	Json(w, r, http.StatusOK, response)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Get("/blocks/after", getBlockAfter)
	return
}
