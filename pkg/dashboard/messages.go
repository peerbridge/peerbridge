package dashboard

import "github.com/peerbridge/peerbridge/pkg/peer"

func reactToPeerMessage(bytes []byte) {
	Hub.broadcast <- bytes
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

	peer.Service.SubscribeIncoming(channel)
	peer.Service.SubscribeOutgoing(channel)
}
