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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "View details about PeerBridge nodes",
	Long:  "Display details of a node inside the PeerBridge blockchain.",
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "View the current account balance of a node",
	Long:  "Retrieve the account balance of a node inside the PeerBridge blockchain.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		msg := fmt.Sprintf("Checking account balance for key %s on host %s.", color.Sprintf(key, color.Notice), color.Sprintf(host, color.Info))
		fmt.Println(msg)

		b, err := GetBalance(host, key)
		if err != nil {
			return fmt.Errorf("Failed to request account balance. %s", err.Error())
		}

		msg = fmt.Sprintf("Account balance: %d", b)
		fmt.Println(msg)

		return
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(balanceCmd)

	nodeCmd.PersistentFlags().StringVar(&key, "key", "", "secp256k1 key of the account")
	nodeCmd.PersistentFlags().StringVar(&host, "host", "https://peerbridge.herokuapp.com", "blockchain node to connect to")

	viper.BindPFlag("key", nodeCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("host", nodeCmd.PersistentFlags().Lookup("host"))

	nodeCmd.MarkPersistentFlagRequired("key")
}

// TODO: move into pkg
func GetBalance(host, key string) (b int64, err error) {
	var pk string

	if len(key) == hex.EncodedLen(secp256k1.PublicKeyByteLength) {
		pk = key
	} else if len(key) == hex.EncodedLen(secp256k1.PrivateKeyByteLength) {
		fmt.Println("Generating public key.")
		kpair, err := secp256k1.LoadKeyPairFromPrivateKeyString(key)
		if err != nil {
			return -1, err
		}
		pk = kpair.PublicKey
	} else {
		return -1, fmt.Errorf("Invalid key format")
	}

	url := fmt.Sprintf("%s/blockchain/accounts/balance/get?account=%s", host, pk)

	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()

	var p blockchain.GetAccountBalanceResponse
	err = json.NewDecoder(res.Body).Decode(&p)
	if err != nil {
		return
	}

	return *p.Balance, nil
}
