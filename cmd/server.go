/*
Copyright © 2021 PeerBridge

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/dashboard"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	. "github.com/peerbridge/peerbridge/pkg/http"
	"github.com/peerbridge/peerbridge/pkg/peer"
	"github.com/peerbridge/peerbridge/pkg/staticfiles"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sync bool

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a new blockchain node",
	Long:  "Start a new PeerBridge blockchain node on the current host",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		kpair, err := secp256k1.LoadKeyPairFromPrivateKeyString(key)
		if err != nil {
			return
		}

		remote := ""
		if sync {
			remote = host
		}

		// Create a http router and start serving http requests
		router := NewRouter()
		router.Use(Header, Logger)

		// Create and run a peer to peer service
		go peer.Service.Run(remote)
		// Bind the peer routes to the main http router
		router.Mount("/peer", peer.Routes())

		// Initiate the blockchain and peer to peer service
		blockchain.InitRepo()
		blockchain.InitChain(kpair)
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

		return
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVar(&key, "key", "", "secp256k1 key of the account")
	serverCmd.PersistentFlags().StringVar(&host, "host", "https://peerbridge.herokuapp.com", "blockchain node to connect to")

	serverCmd.Flags().BoolVar(&sync, "sync", false, "sync the server against the specified host (default is https://peerbridge.herokuapp.com)")

	viper.BindPFlag("key", serverCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("host", serverCmd.PersistentFlags().Lookup("host"))

	serverCmd.MarkPersistentFlagRequired("key")
}
