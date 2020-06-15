package encryption

import "fmt"

func ExampleHybridEncryption() {
	// Alice wants to send this text to bob, with hybrid encryption.
	message := "Incroyable"
	messageData := []byte(message)

	// Both parties need their public/private keypairs first.
	alicePrivateKey, alicePublicKey := CreateRandomAsymetricKeyPair()
	bobPrivateKey, bobPublicKey := CreateRandomAsymetricKeyPair()

	// Alice creates a session key for symmetric encryption.
	sessionKey := CreateRandomSymmetricKey()

	// Alice makes three steps:
	// 1. Encrypt the session key (asymmetrically) with bob's public key
	// 2. Encrypt the message (symmetrically) with the session key
	// 3. Sign the message with her private key
	encryptedSessionKeyHash, encryptedSessionKey := EncryptAsymetrically(
		sessionKey[:],
		bobPublicKey,
	)
	encryptedMessage := EncryptSymmetrically(messageData, sessionKey)
	messageHash, aliceSignature := SignData(messageData, alicePrivateKey)

	// Because none of this data leaks alice's private key,
	// alice can now transmit all of this to bob. When bob gets the data,
	// he does the following steps:
	// 1. Decrypt the symmetric session key with his private key
	// 2. Decrypt the message using the decrypted session key
	// 3. Verify alice's signature using her public key
	decryptedSessionKeySlice := DecryptAsymmetrically(
		encryptedSessionKey,
		encryptedSessionKeyHash,
		bobPrivateKey,
	)
	var decryptedSessionKey [AES256KeySize]byte
	copy(decryptedSessionKey[:], decryptedSessionKeySlice)

	if decryptedSessionKey == sessionKey {
		fmt.Println("Bob was able to reconstruct the symmetric session key.")
	}

	reconstructedMessage := DecryptSymmetrically(encryptedMessage, decryptedSessionKey)
	fmt.Println(fmt.Sprintf("Bob reconstructed the following message: %s", reconstructedMessage))

	err := VerifySignature(reconstructedMessage, alicePublicKey, messageHash, aliceSignature)
	if err == nil {
		fmt.Println("Bob could verify alice's signature.")
	}

	// Output:
	// Bob was able to reconstruct the symmetric session key.
	// Bob reconstructed the following message: Incroyable
	// Bob could verify alice's signature.
}
