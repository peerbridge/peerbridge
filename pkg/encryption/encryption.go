package encryption

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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

func EncryptSymmetrically(data []byte, key AES256Key) (cipherData []byte) {
	block, _ := aes.NewCipher(key[:])
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	cipherData = gcm.Seal(nonce, nonce, data, nil)
	return
}

func DecryptSymmetrically(cipherData []byte, key AES256Key) (data []byte) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, cipher := cipherData[:nonceSize], cipherData[nonceSize:]
	data, err = gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		panic(err.Error())
	}
	return
}

const RSAKeyBitSize = 2048

func CreateRandomAsymetricKeyPair() (privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, RSAKeyBitSize)
	if err != nil {
		panic(err.Error())
	}
	publicKey = &privateKey.PublicKey
	return
}

func EncryptAsymetrically(data []byte, publicKey *rsa.PublicKey) (cipherHash hash.Hash, cipherData []byte) {
	label := []byte("")
	cipherHash = sha256.New()
	cipherData, err := rsa.EncryptOAEP(
		cipherHash,
		rand.Reader,
		publicKey,
		data,
		label,
	)
	if err != nil {
		panic(err.Error())
	}
	return
}

func DecryptAsymmetrically(cipherData []byte, hash hash.Hash, privateKey *rsa.PrivateKey) (data []byte) {
	label := []byte("")
	data, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		privateKey,
		cipherData,
		label,
	)
	if err != nil {
		panic(err.Error())
	}
	return
}

func getSignatureOpts() (opts rsa.PSSOptions) {
	opts.SaltLength = rsa.PSSSaltLengthAuto
	return
}

func SignData(data []byte, privateKey *rsa.PrivateKey) (hashedData []byte, signature []byte) {
	opts := getSignatureOpts()
	hash := crypto.SHA256.New()
	hash.Write(data)
	hashedData = hash.Sum(nil)
	signature, err := rsa.SignPSS(
		rand.Reader,
		privateKey,
		crypto.SHA256,
		hashedData,
		&opts,
	)
	if err != nil {
		panic(err.Error())
	}
	return
}

func VerifySignature(data []byte, publicKey *rsa.PublicKey, hashedData []byte, signature []byte) (err error) {
	opts := getSignatureOpts()
	err = rsa.VerifyPSS(publicKey, crypto.SHA256, hashedData, signature, &opts)
	return
}
