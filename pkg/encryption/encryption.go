package encryption

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// Utilize this function as a reference on how private keys
// should be converted to the PEM string format.
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

// Utilize this function as a reference on how private key
// PEM strings should be converted to private keys.
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

// Utilize this function as a reference on how public keys
// should be converted to the PEM string format.
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

// Utilize this function as a reference on how public key
// PEM strings should be converted to private keys.
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
