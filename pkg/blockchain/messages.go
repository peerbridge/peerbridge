package blockchain

import (
	"encoding/json"
	"log"

	"github.com/peerbridge/peerbridge/pkg/encryption"
	"github.com/peerbridge/peerbridge/pkg/peer"
)

type NewTransactionMessage struct {
	NewTransaction *Transaction `json:"newTransaction"`
}

type NewBlockMessage struct {
	NewBlock *Block `json:"newBlock"`
}

type ResolveBlockResponse struct {
	ResolvedBlock *Block `json:"resolvedBlock"`
}

type ResolveBlockRequest struct {
	BlockID *encryption.SHA256HexString `json:"blockID"`
}

func BroadcastNewTransaction(t *Transaction) {
	log.Printf("Broadcast new transaction: %s\n", t.ID[:6])
	go peer.Instance.Broadcast(NewTransactionMessage{t})
}

func BroadcastNewBlock(b *Block) {
	log.Printf("Broadcast new block: %s\n", b.ID[:6])
	go peer.Instance.Broadcast(NewBlockMessage{b})
}

func BroadcastResolveBlockRequest(id *encryption.SHA256HexString) {
	log.Printf("Broadcast resolve block request: %s\n", (*id)[:6])
	go peer.Instance.Broadcast(ResolveBlockRequest{id})
}

func BroadcastResolveBlockResponse(b *Block) {
	log.Printf("Broadcast resolve block response: %s\n", b.ID[:6])
	go peer.Instance.Broadcast(ResolveBlockResponse{b})
}

func reactToPeerMessage(bytes []byte) {
	log.Println("Got new blockchain peer message.")

	var newTMessage NewTransactionMessage
	err := json.Unmarshal(bytes, &newTMessage)
	if err == nil && newTMessage.NewTransaction != nil {
		Instance.ThreadSafe(func() {
			Instance.AddPendingTransaction(newTMessage.NewTransaction)
		})
		return
	}

	var newBMessage NewBlockMessage
	err = json.Unmarshal(bytes, &newBMessage)
	if err == nil && newBMessage.NewBlock != nil {
		Instance.ThreadSafe(func() {
			Instance.MigrateBlock(newBMessage.NewBlock, false)
		})
		return
	}

	var rRequest ResolveBlockRequest
	err = json.Unmarshal(bytes, &rRequest)
	if err == nil && rRequest.BlockID != nil {
		Instance.ThreadSafe(func() {
			block, err := Instance.GetBlockByID(*rRequest.BlockID)
			if err == nil {
				BroadcastResolveBlockResponse(block)
			}
		})
		return
	}

	var rResponse ResolveBlockResponse
	err = json.Unmarshal(bytes, &rResponse)
	if err == nil && rResponse.ResolvedBlock != nil {
		Instance.ThreadSafe(func() {
			Instance.MigrateBlock(rResponse.ResolvedBlock, false)
		})
		return
	}
}

// Bind the blockchain to new messages from the peer.
func ReactToPeerMessages() {
	channel := make(chan []byte)

	go func() {
		for {
			select {
			case message := <-channel:
				reactToPeerMessage(message)
			}
		}
	}()

	peer.Instance.Subscribe(channel)
}
