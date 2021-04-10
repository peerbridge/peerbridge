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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var amount uint64
var sender string
var receiver string

var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Manage transactions inside the blockchain",
	Long:  "Manage transactions inside the PeerBridge blockchain.",
}

var createTransactionCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new transactions inside the blockchain",
	Long:  "Create new transactions and submit them to the PeerBridge blockchain.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		msg := fmt.Sprintf("Checking account balance for key %s on host %s.", color.Sprintf(sender, color.Notice), color.Sprintf(host, color.Info))
		fmt.Println(msg)

		kpair, err := secp256k1.LoadKeyPairFromPrivateKeyString(sender)
		if err != nil {
			return
		}

		b, err := GetBalance(host, kpair.PublicKey)
		if err != nil {
			return fmt.Errorf("Failed to request account balance. %s", err.Error())
		}

		msg = fmt.Sprintf("Account balance: %d", b)
		fmt.Println(msg)

		if int64(amount) > b {
			return errors.New("Failed to create transaction. Amount exceeds account balance!")
		}

		err = createTransaction(host, amount, kpair, receiver)
		if err != nil {
			return fmt.Errorf("Failed to create transaction. %s", err.Error())
		}

		msg = fmt.Sprintf(
			"Create new transaction containing %s coins from %s to %s on host %s.",
			color.Sprintf(fmt.Sprint(amount), color.Debug),
			color.Sprintf(sender, color.Notice),
			color.Sprintf(receiver, color.Notice),
			color.Sprintf(host, color.Info),
		)
		fmt.Println(msg)

		return
	},
}

func init() {
	rootCmd.AddCommand(transactionCmd)
	transactionCmd.AddCommand(createTransactionCmd)

	transactionCmd.PersistentFlags().StringVar(&host, "host", "https://peerbridge.herokuapp.com", "blockchain node to connect to")

	viper.BindPFlag("host", transactionCmd.PersistentFlags().Lookup("host"))

	createTransactionCmd.Flags().Uint64Var(&amount, "amount", uint64(0), "Amount to transfer as part of the transaction")
	createTransactionCmd.Flags().StringVar(&sender, "sender", "", "secp256k1 private key of the account to create a transaction")
	createTransactionCmd.Flags().StringVar(&receiver, "receiver", "", "secp256k1 public key of the receiver of the transaction")

	createTransactionCmd.MarkFlagRequired("amount")
	createTransactionCmd.MarkFlagRequired("sender")
	createTransactionCmd.MarkFlagRequired("receiver")
}

func createTransaction(host string, amount uint64, kpair *secp256k1.KeyPair, receiver string) (err error) {
	randomID, err := encryption.RandomSHA256HexString()
	if err != nil {
		return
	}

	t := &blockchain.Transaction{
		ID:           *randomID,
		Sender:       kpair.PublicKey,
		Receiver:     receiver,
		Balance:      amount,
		TimeUnixNano: time.Now().UnixNano(),
		Data:         nil,
		Fee:          0,
		Signature:    nil, // part of signing
	}

	signature, err := secp256k1.ComputeSignature(t, kpair.PrivateKey)
	if err != nil {
		return
	}

	t.Signature = signature

	data, err := json.Marshal(blockchain.CreateTransactionRequest{Transaction: t})
	if err != nil {
		return
	}

	url := fmt.Sprintf("%s/blockchain/transaction/create", host)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Println(res.StatusCode)
		err = errors.New("Something went wrong!")
		return
	}

	for {
		url = fmt.Sprintf("%s/blockchain/transaction/get?id=%s", host, t.ID)
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		res.Body.Close()

		if res.StatusCode == http.StatusOK {
			return nil
		} else if res.StatusCode == http.StatusAccepted {
			time.Sleep(1 * time.Second)
		}

	}
}
