package encryption

import (
	"encoding/json"
	"log"
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

type CreateAsymmetricKeyPairResponse struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

func handleError(err error, w http.ResponseWriter) {
	log.Println(err.Error())
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Create an asymmetric key pair.
//
// TODO: Do this on the client.
func createAsymmetricKeyPair(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	keyPair, err := CreateRandomAsymmetricKeyPair()
	publicKeyString, err := PublicKeyToPEMString(keyPair.PublicKey)
	if err != nil {
		handleError(err, w)
		return
	}
	privateKeyString := PrivateKeyToPEMString(keyPair.PrivateKey)
	w.WriteHeader(http.StatusCreated)
	response := CreateAsymmetricKeyPairResponse{privateKeyString, *publicKeyString}
	json.NewEncoder(w).Encode(response)
}

type CreateSymmetricKeyResponse struct {
	Key [32]byte `json:"key"`
}

// Create a symmetric key.
//
// TODO: Do this on the client.
func createSymmetricKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := CreateRandomSymmetricKey()
	w.WriteHeader(http.StatusCreated)
	response := CreateSymmetricKeyResponse{key}
	json.NewEncoder(w).Encode(response)
}

var Routes = []Route{
	Route{Method: http.MethodPost, Pattern: "/credentials/asymmetric/new", Handler: http.HandlerFunc(createAsymmetricKeyPair)},
	Route{Method: http.MethodPost, Pattern: "/credentials/symmetric/new", Handler: http.HandlerFunc(createSymmetricKey)},
}
