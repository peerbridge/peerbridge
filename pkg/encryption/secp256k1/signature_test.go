package secp256k1

import (
	"encoding/hex"
	"log"
	"testing"

	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var message = "Hello"
var privKey = "60f8700baf057e6131b912b97f2e36f54a67544a5f4659de348e988306ab1a3f"
var pubKey = "02caa8bded7764cca5bde64c10ae54fc91f4bcd2de08eb4c66b1e2dc3d9dd5519d"
var signature = "b97676fd0290f9b98c02e0bf11c495af25fd4126a63e46ec90907c15ae7ff30a727e97bf424f31d067becdcc5a0549441e9a918bbe6defc2dcaea52021bab267"

func TestSign(t *testing.T) {
	i := NewSigningInput([]byte(message))
	k := PrivateKeyHexString(privKey)
	s, err := i.Sign(k)
	if err != nil {
		t.Fatal(err)
	}

	if *s == signature {
		log.Printf("Signature matches: %s", *s)
	} else {
		t.Errorf("Expected signature to match")
	}
}

func TestVerifySignature(t *testing.T) {
	sig, err := hex.DecodeString(signature)
	if err != nil {
		t.Fatal(err)
	}

	pk, err := hex.DecodeString(pubKey)
	if err != nil {
		t.Fatal(err)
	}

	i := NewSigningInput([]byte(message))

	if ok := ethsecp256k1.VerifySignature(pk, i.Bytes[:], sig); ok {
		log.Println("Signature Valid")
	} else {
		t.Error("Expected signature to be valid")
	}
}
