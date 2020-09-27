package main

import (
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
	models := []interface{}{
		(*blockchain.Block)(nil),
		(*blockchain.Transaction)(nil),
	}

	err := database.Initialize(models)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(blockCreationInterval * time.Second)
	go blockchain.ScheduleBlockCreation(ticker)

	router := NewRouter()
	router.Use(Header, Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to PeerBridge!"))
	})
	router.Mount("/blockchain", blockchain.Routes())

	fmt.Println(fmt.Sprintf("Start server listening on: %s", color.Sprintf(GetServerPort(), color.Info)))
	log.Fatal(router.ListenAndServe())
}
