package http

import (
	"fmt"
	"net/http"
	"os"
)

const DefaultPort = "8000"

func getServerPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return DefaultPort
}

func NewServer() *http.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/block/new", newBlock)
	handler.HandleFunc("/block/hash", hashBlock)

	addr := fmt.Sprintf(":%s", getServerPort())
	return &http.Server{Addr: addr, Handler: handler}
}
