package encryption

import (
	"net/http"

	. "github.com/peerbridge/peerbridge/pkg/http"
)

type CreateAsymmetricKeyPairResponse struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

// Create an asymmetric key pair.
//
// TODO: Do this on the client.
func createAsymmetricKeyPair(w http.ResponseWriter, r *http.Request) {
	keyPair, err := CreateRandomAsymmetricKeyPair()
	publicKeyString, err := PublicKeyToPEMString(keyPair.PublicKey)
	if err != nil {
		InternalServerError(w, err)
		return
	}
	privateKeyString := PrivateKeyToPEMString(keyPair.PrivateKey)
	response := CreateAsymmetricKeyPairResponse{privateKeyString, *publicKeyString}
	Json(w, r, http.StatusCreated, response)
}

type CreateSymmetricKeyResponse struct {
	Key [32]byte `json:"key"`
}

// Create a symmetric key.
//
// TODO: Do this on the client.
func createSymmetricKey(w http.ResponseWriter, r *http.Request) {
	key := CreateRandomSymmetricKey()
	response := CreateSymmetricKeyResponse{key}
	Json(w, r, http.StatusCreated, response)
}

func Routes() (router *Router) {
	router = NewRouter()
	router.Post("/asymmetric/new", createAsymmetricKeyPair)
	router.Post("/symmetric/new", createSymmetricKey)
	return
}
