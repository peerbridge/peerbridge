package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/dashboard"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	. "github.com/peerbridge/peerbridge/pkg/http"
	"github.com/peerbridge/peerbridge/pkg/peer"
	"github.com/peerbridge/peerbridge/pkg/staticfiles"
)

func main() {
	// Get the needed environment variables
	var keyPair *secp256k1.KeyPair
	privateKeyString := os.Getenv("PRIVATE_KEY")
	if privateKeyString != "" {
		// Load the keypair from the private key string
		var err error
		keyPair, err = secp256k1.LoadKeyPairFromPrivateKeyString(privateKeyString)
		if err != nil {
			panic(err)
		}
	} else {
		// Load the keypair or store a new one
		keyPath := os.Getenv("KEY_PATH")
		if keyPath == "" {
			keyPath = "./key.json" // default key path
		}

		var err error
		keyPair, err = secp256k1.LoadKeyPair(keyPath)
		if err != nil {
			keyPair, err = secp256k1.StoreNewKeyPair(keyPath)
			if err != nil {
				panic(err)
			}
		}
	}

	remote := os.Getenv("REMOTE_URL")
	if remote == "" {
		log.Println(color.Sprintf("No REMOTE_URL set. Will not bootstrap this service", color.Warning))
	}

	// Create a http router and start serving http requests
	router := NewRouter()
	router.Use(Header, Logger)

	// Create and run a peer to peer service
	go peer.Instance.Run(remote)
	// Bind the peer routes to the main http router
	router.Mount("/peer", peer.Routes())

	// Initiate the blockchain and peer to peer service
	blockchain.Init(keyPair)
	blockchain.Instance.Sync(remote)
	go blockchain.ReactToPeerMessages()
	go blockchain.Instance.RunContinuousMinting()
	// Bind the blockchain routes to the main http router
	router.Mount("/blockchain", blockchain.Routes())

	// Run the dashboard websocket client hub
	go dashboard.RunHub()
	go dashboard.ReactToPeerMessages()
	// Bind the dashboard routes to the main http router
	router.Mount("/dashboard", dashboard.Routes())

	// Bind the staticfiles routes to the main http router
	router.Mount("/static", staticfiles.Routes())

	// Redirect index page visits to the dashboard
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", 301)
	})

	// Finish initiation and listen for requests
	log.Println(fmt.Sprintf("Started http server listening on: %s", color.Sprintf(GetServerPort(), color.Info)))
	log.Fatal(router.ListenAndServe())
}
