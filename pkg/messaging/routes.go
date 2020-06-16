package messaging

import (
	"encoding/json"
	"net/http"

	"github.com/peerbridge/peerbridge/pkg/blockchain"
	"github.com/peerbridge/peerbridge/pkg/encryption"
	. "github.com/peerbridge/peerbridge/pkg/http"
)

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
	var b SendMessageRequest

	err := DecodeJSONBody(w, r, &b)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	decryptedPublicKey, err := encryption.PEMStringToPublicKey(b.ReceiverPublicKey)
	decryptedPrivateKey, err := encryption.PEMStringToPrivateKey(b.PrivateKey)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	encryptedSessionKey, err := encryption.EncryptAsymmetrically(
		b.SessionKey[:],
		decryptedPublicKey,
	)
	if err != nil {
		InternalServerError(w, err)
		return
	}

	messageData := []byte(b.Content)
	encryptedMessage, err := encryption.EncryptSymmetrically(messageData, b.SessionKey)
	if err != nil {
		InternalServerError(w, err)
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
	Json(w, r, http.StatusCreated, response)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/new", sendMessage)
	return
}
