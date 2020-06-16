package main

import (
	"fmt"
	"log"
	"time"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/http"
	"github.com/peerbridge/peerbridge/pkg/messaging"
)

const blockCreationInterval = 3

func main() {
	ticker := time.NewTicker(blockCreationInterval * time.Second)
	go blockchain.ScheduleBlockCreation(ticker)

	router := http.NewRouter()
	router.Use(http.Header, http.Logger)
	router.Mount("/credentials", encryption.Routes())
	router.Mount("/messages", messaging.Routes())

	fmt.Println(fmt.Sprintf("Start server listening on: %s", color.Sprintf(http.GetServerPort(), color.Info)))
	log.Fatal(router.ListenAndServe())
}
