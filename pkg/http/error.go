package http

import (
	"log"
	"net/http"
)

// Dispatch an internal server error response using the http
// response writer.
func InternalServerError(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
