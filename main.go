package main

import (
	"fmt"
	"log"

	"github.com/peerbridge/peerbridge/pkg/http"
)

func main() {
	server := http.NewServer()
	fmt.Println(fmt.Sprintf("Start server listening on: %s", server.Addr))
	log.Fatal(server.ListenAndServe())
}
