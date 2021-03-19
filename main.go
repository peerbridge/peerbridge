package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/database"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	. "github.com/peerbridge/peerbridge/pkg/http"
)

func main() {
	keypath := flag.
		String("k", "", "The path to the private key.")
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

	if keypath == nil || *keypath == "" {
		panic("Keypath must be supplied via -k!")
	}

	keyPair, err := secp256k1.LoadKeyPair(*keypath)
	if err != nil {
		keyPair, err = secp256k1.StoreNewKeyPair(*keypath)
		if err != nil {
			panic(err)
		}
	}

	go blockchain.Peer.Run(remote)

	blockchain.Init(keyPair)
	go blockchain.Instance.RunContinuousMinting()

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
