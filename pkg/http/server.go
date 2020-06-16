package http

import (
	"fmt"
	"net/http"
	"os"
)

const DefaultPort = "8080"

func getServerPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return DefaultPort
}

func CreateServer(router Router) *http.Server {
	// generate server address
	addr := fmt.Sprintf(":%s", getServerPort())

	return &http.Server{Addr: addr, Handler: router}
}
