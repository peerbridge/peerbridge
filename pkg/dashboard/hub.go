package dashboard

// Hub maintains the set of active clients and
// broadcasts messages to the clients.
type HubService struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

var Hub = &HubService{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func RunHub() {
	for {
		select {
		case client := <-Hub.register:
			Hub.clients[client] = true
		case client := <-Hub.unregister:
			if _, ok := Hub.clients[client]; ok {
				delete(Hub.clients, client)
				close(client.send)
			}
		case message := <-Hub.broadcast:
			for client := range Hub.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(Hub.clients, client)
				}
			}
		}
	}
}

func Publish(message []byte) {
	Hub.broadcast <- message
}
