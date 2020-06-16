package blockchain

import "fmt"

func ExampleBlockchainExtension() {
	blockChain := BlockChain{
		[]Block{},
		[]Transaction{},
	}

	alicePublicKeyString := "This is a fictitious public key string for alice."
	bobPublicKeyString := "This is a fictitious public key string for bob."
	message1Data := []byte("Hi alice, I send you a cleartext string.")
	message2Data := []byte("Hi bob, but what if you want to send more than that?")
	message3Data := []byte("You can send anything serializable, really.")
	message4Data := []byte("I see. Let's use json data with encryption information from now on.")

	transaction1 := Transaction{bobPublicKeyString, alicePublicKeyString, message1Data}
	blockChain.AddTransaction(transaction1)
	transaction2 := Transaction{alicePublicKeyString, bobPublicKeyString, message2Data}
	blockChain.AddTransaction(transaction2)

	blockChain.ForgeNewBlock()
	fmt.Println("Forged the first block. This many transactions are to be found for Alice:")
	fmt.Println(len(blockChain.GetForgedTransactions(alicePublicKeyString)))

	transaction3 := Transaction{bobPublicKeyString, alicePublicKeyString, message3Data}
	blockChain.AddTransaction(transaction3)

	blockChain.ForgeNewBlock()
	fmt.Println("Forged the second block. This many transactions are to be found for Alice:")
	fmt.Println(len(blockChain.GetForgedTransactions(alicePublicKeyString)))

	blockChain.ForgeNewBlock()
	fmt.Println("Forged the third block. This many transactions are to be found for Alice:")
	fmt.Println(len(blockChain.GetForgedTransactions(alicePublicKeyString)))

	transaction4 := Transaction{alicePublicKeyString, bobPublicKeyString, message4Data}
	blockChain.AddTransaction(transaction4)

	blockChain.ForgeNewBlock()
	fmt.Println("Forged the fourth block. This many transactions are to be found for Alice:")
	fmt.Println(len(blockChain.GetForgedTransactions(alicePublicKeyString)))

	// Output:
	// Forged the first block. This many transactions are to be found for Alice:
	// 2
	// Forged the second block. This many transactions are to be found for Alice:
	// 3
	// Forged the third block. This many transactions are to be found for Alice:
	// 3
	// Forged the fourth block. This many transactions are to be found for Alice:
	// 4
}
