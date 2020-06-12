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

func NewServer() *http.Server {
	// create a new Router to handle http request
	router := NewRouter()

	// add Middleware to the router
	router.Use(header)
	router.Use(logger)

	// register routes
	router.Add(Routes...)

	// generate server address
	addr := fmt.Sprintf(":%s", getServerPort())

	return &http.Server{Addr: addr, Handler: router}
}
