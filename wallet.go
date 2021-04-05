package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/color"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/encryption/secp256k1"
)

// An example implementation of a simple wallet.
func Run() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to the basic peerbridge wallet!")
	fmt.Printf("Please authenticate with your private key: ")

	enteredPrivateKeyString, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	enteredPrivateKeyString = strings.ToLower(strings.TrimSpace(enteredPrivateKeyString))

	// Load the keypair from the private key string
	keyPair, err := secp256k1.LoadKeyPairFromPrivateKeyString(enteredPrivateKeyString)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Please enter the remote you want to contact: ")

	enteredRemote, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	enteredRemote = strings.ToLower(strings.TrimSpace(enteredRemote))

	fmt.Println("Checking your account balance...")
	url := fmt.Sprintf("%s/blockchain/accounts/balance/get?account=%s", enteredRemote, keyPair.PublicKey)
	body := bytes.NewBuffer([]byte{})
	request, err := http.NewRequest("GET", url, body)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var accountResponse blockchain.GetAccountBalanceResponse
	err = json.Unmarshal(responseBody, &accountResponse)
	if err != nil {
		panic(err)
	}

	response.Body.Close()

	fmt.Printf("Your account balance: %d\n", *accountResponse.Balance)

	if *accountResponse.Balance <= 0 {
		fmt.Println("You're broke. You cannot send money.")
		return
	}

	fmt.Printf("The receiver for your transaction: ")

	enteredReceiverPublicKeyString, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	enteredReceiverPublicKeyString = strings.ToLower(strings.TrimSpace(enteredReceiverPublicKeyString))

	fmt.Printf("The amount you want to send: ")

	enteredAmountString, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	enteredAmount, err := strconv.Atoi(strings.ToLower(strings.TrimSpace(enteredAmountString)))
	if err != nil {
		panic(err)
	}

	if int64(enteredAmount) > *accountResponse.Balance {
		panic("You cannot send that much!")
	}
	if enteredAmount < 0 {
		panic("You cannot send negative balances!")
	}

	randomID, err := encryption.RandomSHA256HexString()
	if err != nil {
		panic(err)
	}
	t := &blockchain.Transaction{
		ID:           *randomID,
		Sender:       keyPair.PublicKey,
		Receiver:     enteredReceiverPublicKeyString,
		Balance:      uint64(enteredAmount),
		TimeUnixNano: time.Now().UnixNano(),
		Data:         nil,
		Fee:          0,
		Signature:    nil, // part of signing
	}
	signature, err := secp256k1.ComputeSignature(t, keyPair.PrivateKey)
	if err != nil {
		panic(err)
	}
	t.Signature = signature

	fmt.Println("Sending a new transaction...")
	url = fmt.Sprintf("%s/blockchain/transaction/create", enteredRemote)
	requestData := blockchain.CreateTransactionRequest{Transaction: t}
	tBytes, err := json.Marshal(requestData)
	if err != nil {
		panic(err)
	}
	body = bytes.NewBuffer(tBytes)
	request, err = http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err = client.Do(request)
	if err != nil {
		panic(err)
	}

	if response.StatusCode != 200 {
		panic("Something went wrong!")
	}

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	response.Body.Close()

	fmt.Println("Transaction successfully submitted!")
	fmt.Println("-----------------------------------")
	fmt.Println("Waiting for the transaction to be included...")

	for {
		url = fmt.Sprintf("%s/blockchain/transaction/get?id=%s", enteredRemote, t.ID)
		body = bytes.NewBuffer([]byte{})
		request, err = http.NewRequest("GET", url, body)
		if err != nil {
			panic(err)
		}
		request.Header.Set("Content-Type", "application/json")

		response, err = client.Do(request)
		if err != nil {
			panic(err)
		}

		responseBody, err = ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}

		response.Body.Close()

		if response.StatusCode == 202 {
			fmt.Printf("Transaction status: %s...\n", color.Sprintf("Pending", color.Warning))
			time.Sleep(time.Second * 1)
		}
		if response.StatusCode == 200 {
			fmt.Printf("Transaction status: %s\n", color.Sprintf("Included in blockchain!", color.Success))
			break
		}
	}
}
