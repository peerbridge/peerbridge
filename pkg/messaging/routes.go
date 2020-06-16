package messaging

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	. "github.com/peerbridge/peerbridge/pkg/http"
)

func handleError(err error, w http.ResponseWriter) {
	log.Println(err.Error())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

type SendMessageRequest struct {
	PrivateKey        string   `json:"privateKey"`
	PublicKey         string   `json:"publicKey"`
	ReceiverPublicKey string   `json:"receiverPublicKey"`
	SessionKey        [32]byte `json:"sessionKey"`
	Content           string   `json:"content"`
}

type SendMessageResponse struct {
	Transaction blockchain.Transaction `json:"message"`
}

// Send a message to another client.
//
// TODO: Do message encryption and signing on the client.
func sendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var b SendMessageRequest
	var err error

	err = DecodeJSONBody(w, r, &b)
	if err != nil {
		handleError(err, w)
		return
	}

	decryptedPublicKey, err := encryption.PEMStringToPublicKey(b.ReceiverPublicKey)
	decryptedPrivateKey, err := encryption.PEMStringToPrivateKey(b.PrivateKey)
	if err != nil {
		handleError(err, w)
		return
	}

	encryptedSessionKey, err := encryption.EncryptAsymmetrically(
		b.SessionKey[:],
		decryptedPublicKey,
	)
	if err != nil {
		handleError(err, w)
		return
	}

	messageData := []byte(b.Content)
	encryptedMessage, err := encryption.EncryptSymmetrically(messageData, b.SessionKey)
	if err != nil {
		handleError(err, w)
		return
	}

	signatureData, err := encryption.SignData(messageData, decryptedPrivateKey)
	if err != nil {
		return
	}

	message := Message{*signatureData, *encryptedSessionKey, *encryptedMessage}
	messageJsonData, _ := json.Marshal(message)
	transaction := blockchain.Transaction{
		Sender:   b.PublicKey,
		Receiver: b.ReceiverPublicKey,
		Data:     messageJsonData,
	}
	blockchain.MainBlockChain.AddTransaction(transaction)
	response := SendMessageResponse{transaction}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

var Routes = []Route{
	Route{Method: http.MethodPost, Pattern: "/messages/new", Handler: http.HandlerFunc(sendMessage)},
}
