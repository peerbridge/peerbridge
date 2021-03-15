package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

func AliceExamplePublicKey() string {
	key, err := LoadPrivateKey("./alice.priv.key")
	if err != nil {
		panic(err)
	}
	return PublicKeyToPEMString(&key.PublicKey)
}

func BobExamplePublicKey() string {
	key, err := LoadPrivateKey("./bob.priv.key")
	if err != nil {
		panic(err)
	}
	return PublicKeyToPEMString(&key.PublicKey)
}

func LoadPrivateKey(keypath string) (*rsa.PrivateKey, error) {
	bytes, err := ioutil.ReadFile(keypath)
	if err != nil {
		return nil, err
	}
	key, err := PEMStringToPrivateKey(string(bytes))
	if err != nil {
		return nil, err
	}
	return key, nil
}

func StoreNewPrivateKey(keypath string) (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	bytes := []byte(PrivateKeyToPEMString(key))
	err = ioutil.WriteFile(keypath, bytes, 0644)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func PrivateKeyToPEMString(privateKey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	return string(privateKeyPEM)
}

func PEMStringToPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("Failed to parse PEM private key block.")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func PublicKeyToPEMString(publicKey *rsa.PublicKey) string {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyPEMBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)
	publicKeyPEM := string(publicKeyPEMBytes)
	return publicKeyPEM
}

func PEMStringToPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("Failed to parse PEM public key block.")
	}
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}
