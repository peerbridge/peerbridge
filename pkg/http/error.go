package http

import (
	"log"
	"net/http"
)

func InternalServerError(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
