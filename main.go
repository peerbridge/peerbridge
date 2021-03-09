package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/database"
	. "github.com/peerbridge/peerbridge/pkg/http"
)

const blockCreationInterval = 3

func main() {
	bootstrapTarget := flag.
		String("bootstrap", "", "The bootstrap target url")
	flag.Parse()

	// Initialize the database models
	models := []interface{}{
		(*blockchain.Block)(nil),
		(*blockchain.Transaction)(nil),
	}
	err := database.Initialize(models)
	if err != nil {
		panic(err)
	}

	// Run the p2p peer server concurrently
	go blockchain.P2PServiceInstance.Run(bootstrapTarget)

	// Schedule the periodic block creation
	ticker := time.NewTicker(blockCreationInterval * time.Second)
	go blockchain.ScheduleBlockCreation(ticker)

	// Create a http router and start serving http requests
	router := NewRouter()
	router.Use(Header, Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to PeerBridge!"))
	})
	router.Mount("/blockchain", blockchain.Routes())

	fmt.Println(fmt.Sprintf("Start server listening on: %s", color.Sprintf(GetServerPort(), color.Info)))
	log.Fatal(router.ListenAndServe())
}
