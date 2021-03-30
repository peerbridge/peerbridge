package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	. "github.com/peerbridge/peerbridge/pkg/http"
)

func main() {
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

	blockchain.Init(keyPair)                   // blocking
	blockchain.Instance.Sync(remote)           // blocking
	blockchain.Peer.Run(&remote)               // concurrent
	blockchain.Instance.RunContinuousMinting() // concurrent

	// Create a http router and start serving http requests
	router := NewRouter()
	router.Use(Header, Logger)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to PeerBridge!"))
	})
	router.Mount("/blockchain", blockchain.Routes())

	log.Println(fmt.Sprintf("Start REST server listening on: %s", color.Sprintf(GetServerPort(), color.Info)))
	log.Fatal(router.ListenAndServe())
}
