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
	remote := flag.
		String("r", "", "A remote for P2P bootstrapping and catching up.")
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

	go peer.Instance.Run(remote)

	blockchain.Init(key)
	go blockchain.Instance.CatchUp(remote, func() {
		go blockchain.Instance.RunContinuousMinting()
		go blockchain.Instance.ListenOnRemoteUpdates()
	})

	// Create a http router and start serving http requests
	router := NewRouter()
	router.Use(Header, Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to PeerBridge!"))
	})
	router.Mount("/peer", peer.Routes())
	router.Mount("/blockchain", blockchain.Routes())

	log.Println(fmt.Sprintf("Start REST server listening on: %s", color.Sprintf(GetServerPort(), color.Info)))
	log.Fatal(router.ListenAndServe())
}
