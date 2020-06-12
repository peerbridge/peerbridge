package main

import (
	"fmt"
	"log"

	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/http"
)

func main() {
	server := http.NewServer()
	fmt.Println(fmt.Sprintf("Start server listening on: %s", color.Sprintf(server.Addr, color.InfoColor)))
	log.Fatal(server.ListenAndServe())
}
