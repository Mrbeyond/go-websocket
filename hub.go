package main

import "encoding/json"

type Hub struct {
	clients    map[*CLient]bool
	broadcast  chan []byte  // Incoming message from a client to other clients
	register   chan *CLient //Channel conveying a new client
	unregister chan *CLient //Channel conveying leaving client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *CLient),
		unregister: make(chan *CLient),
		clients:    make(map[*CLient]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case newClient := <-h.register:
			ID := newClient.User.ID
			for client := range h.clients {
				message := []byte("Someone joined  the room: (Name: " + ID + ")")
				client.send <- message
			}
			h.clients[newClient] = true

		case leavingClient := <-h.unregister:
			ID := leavingClient.User.ID

			if _, ok := h.clients[leavingClient]; ok {
				delete(h.clients, leavingClient)
				close(leavingClient.send)
			}
			for client := range h.clients {
				message := []byte("Someone left  the room: (Name: " + ID + ")")
				client.send <- message
			}
		case userMessage := <-h.broadcast:
			var data map[string][]byte
			json.Unmarshal(userMessage, &data)
			for client := range h.clients {
				//Don't broadcast to self
				if client.User.ID == string(data["id"]) {
					continue
				}
				for {
					select {
					case client.send <- data["message"]:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}

	}
}
