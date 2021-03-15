package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/database"
	. "github.com/peerbridge/peerbridge/pkg/http"
	"github.com/peerbridge/peerbridge/pkg/peer"
)

const blockCreationInterval = 3

func main() {
	bootstrapTarget := flag.
		String("bootstrap", "", "The bootstrap target url. If not given, the node will not attempt to bootstrap in the P2P network.")
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

	key, err := rsa.GenerateKey(rand.Reader, 2048)

	_peer := peer.CreateP2PService()
	go _peer.Run(bootstrapTarget)

	_blockchain := blockchain.CreateNewBlockchain(key)
	go _blockchain.RunContinuousMinting()
	go _blockchain.ListenOnRemoteUpdates()

	// Create a http router and start serving http requests
	router := NewRouter()
	router.Use(Header, Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to PeerBridge!"))
	})

	log.Println(fmt.Sprintf("Start REST server listening on: %s", color.Sprintf(GetServerPort(), color.Info)))
	log.Fatal(router.ListenAndServe())
}
