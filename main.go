package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	. "github.com/peerbridge/peerbridge/pkg/http"
	"github.com/peerbridge/peerbridge/pkg/messaging"
)

var Routes = append(encryption.Routes, messaging.Routes...)

const DefaultPort = "8080"

func getServerPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return DefaultPort
}

func CreateServer() *http.Server {
	// create a new Router to handle http request
	router := NewRouter()

	// add Middleware to the router
	router.Use(Header)
	router.Use(Logger)

	// register routes
	router.Add(Routes...)

	// generate server address
	addr := fmt.Sprintf(":%s", getServerPort())

	return &http.Server{Addr: addr, Handler: router}
}

func ScheduleBlockCreation() {
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for t := range ticker.C {
			if len(blockchain.MainBlockChain.PendingTransactions) == 0 {
				continue
			}
			log.Printf(
				"%s: Forging a new Block. Blocks: %d, Transactions: %d",
				t, len(blockchain.MainBlockChain.Blocks), len(blockchain.MainBlockChain.GetAllForgedTransactions()),
			)
			blockchain.MainBlockChain.ForgeNewBlock()
		}
	}()
}

func main() {
	ScheduleBlockCreation()
	server := CreateServer()
	fmt.Println(fmt.Sprintf("Start server listening on: %s", color.Sprintf(server.Addr, color.InfoColor)))
	log.Fatal(server.ListenAndServe())
}
