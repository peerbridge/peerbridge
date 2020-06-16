package encryption

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
	"io"
)

const AES256KeySize = 32

type AES256Key = [AES256KeySize]byte

func CreateRandomSymmetricKey() (key AES256Key) {
	slice := make([]byte, AES256KeySize)
	rand.Read(slice)
	copy(key[:], slice)
	return
}

func EncryptSymmetrically(data []byte, key AES256Key) (*[]byte, error) {
	block, _ := aes.NewCipher(key[:])
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	cipherData := gcm.Seal(nonce, nonce, data, nil)
	return &cipherData, nil
}

func DecryptSymmetrically(cipherData []byte, key AES256Key) (*[]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, cipher := cipherData[:nonceSize], cipherData[nonceSize:]
	data, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

const RSAKeyBitSize = 2048

type RSAKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func CreateRandomAsymmetricKeyPair() (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, RSAKeyBitSize)
	if err != nil {
		return nil, err
	}
	keyPair := RSAKeyPair{
		privateKey,
		&privateKey.PublicKey,
	}
	return &keyPair, nil
}

type EncryptedData struct {
	CipherHash hash.Hash
	CipherData []byte
}

func EncryptAsymmetrically(data []byte, publicKey *rsa.PublicKey) (*EncryptedData, error) {
	label := []byte("")
	cipherHash := sha256.New()
	cipherData, err := rsa.EncryptOAEP(
		cipherHash,
		rand.Reader,
		publicKey,
		data,
		label,
	)
	if err != nil {
		return nil, err
	}
	return &EncryptedData{cipherHash, cipherData}, nil
}

func DecryptAsymmetrically(cipherData []byte, hash hash.Hash, privateKey *rsa.PrivateKey) (*[]byte, error) {
	label := []byte("")
	data, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		privateKey,
		cipherData,
		label,
	)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func getSignatureOpts() (opts rsa.PSSOptions) {
	opts.SaltLength = rsa.PSSSaltLengthAuto
	return
}

type SignatureData struct {
	hashedData []byte
	signature  []byte
}

func SignData(data []byte, privateKey *rsa.PrivateKey) (*SignatureData, error) {
	opts := getSignatureOpts()
	hash := crypto.SHA256.New()
	hash.Write(data)
	hashedData := hash.Sum(nil)
	signature, err := rsa.SignPSS(
		rand.Reader,
		privateKey,
		crypto.SHA256,
		hashedData,
		&opts,
	)
	if err != nil {
		return nil, err
	}
	return &SignatureData{hashedData, signature}, nil
}

func VerifySignature(data []byte, publicKey *rsa.PublicKey, i SignatureData) (err error) {
	opts := getSignatureOpts()
	err = rsa.VerifyPSS(publicKey, crypto.SHA256, i.hashedData, i.signature, &opts)
	return
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

func PublicKeyToPEMString(publicKey *rsa.PublicKey) (*string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEMBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)
	publicKeyPEM := string(publicKeyPEMBytes)
	return &publicKeyPEM, nil
}

func PEMStringToPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("Failed to parse PEM public key block.")
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break
	}
	return nil, errors.New("The given key is no RSA public key.")
}
