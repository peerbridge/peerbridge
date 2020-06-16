package encryption

import "fmt"

func ExampleHybridEncryption() {
	message := "Incroyable"
	messageData := []byte(message)

	aliceKeyPair, errAlice := CreateRandomAsymmetricKeyPair()
	bobKeyPair, errBob := CreateRandomAsymmetricKeyPair()
	if errAlice != nil || errBob != nil {
		return
	}
	fmt.Println("Alice and Bob successfully created their random keypairs.")

	bobPublicKeyString, errBob := PublicKeyToPEMString(bobKeyPair.PublicKey)
	if errAlice != nil || errBob != nil {
		return
	}
	fmt.Println("Bob published his public key.")

	sessionKey := CreateRandomSymmetricKey()
	fmt.Println("Alice created the random symmetric session key.")

	decodedPublicKeyBob, err := PEMStringToPublicKey(*bobPublicKeyString)
	if err != nil {
		return
	}
	fmt.Println("Alice decrypted bob's public key.")
	encryptedData, err := EncryptAsymmetrically(
		sessionKey[:],
		decodedPublicKeyBob,
	)
	if err != nil {
		return
	}
	fmt.Println("Alice encrypted the session key asymmetrically.")
	encryptedMessage, err := EncryptSymmetrically(messageData, sessionKey)
	if err != nil {
		return
	}
	fmt.Println("Alice encrypted the message data symmetrically.")
	aliceSignatureData, err := SignData(messageData, aliceKeyPair.PrivateKey)
	if err != nil {
		return
	}
	fmt.Println("Alice successfully signed the message data.")

	decryptedSessionKeySlice, err := DecryptAsymmetrically(
		encryptedData.CipherData,
		encryptedData.CipherHash,
		bobKeyPair.PrivateKey,
	)
	if err != nil {
		return
	}
	var decryptedSessionKey [AES256KeySize]byte
	copy(decryptedSessionKey[:], *decryptedSessionKeySlice)

	if decryptedSessionKey == sessionKey {
		fmt.Println("Bob was able to reconstruct the symmetric session key.")
	}

	reconstructedMessage, err := DecryptSymmetrically(*encryptedMessage, decryptedSessionKey)
	if err != nil {
		return
	}
	fmt.Println(fmt.Sprintf("Bob reconstructed the following message: %s", *reconstructedMessage))

	err = VerifySignature(*reconstructedMessage, aliceKeyPair.PublicKey, *aliceSignatureData)
	if err == nil {
		fmt.Println("Bob could verify alice's signature.")
	}

	// Output:
	// Alice and Bob successfully created their random keypairs.
	// Bob published his public key.
	// Alice created the random symmetric session key.
	// Alice decrypted bob's public key.
	// Alice encrypted the session key asymmetrically.
	// Alice encrypted the message data symmetrically.
	// Alice successfully signed the message data.
	// Bob was able to reconstruct the symmetric session key.
	// Bob reconstructed the following message: Incroyable
	// Bob could verify alice's signature.
}
