package main

import "log"

type Message struct {
	Sender *Client
	Text   []byte
}

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	peers map[*Client]*Client

	send chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		send:       make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		peers:      make(map[*Client]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			coupled := false
			for k, v := range h.peers {
				if v == nil {
					h.peers[k] = client
					h.peers[client] = k
					coupled = true
					log.Println("new clients coupled!")
					break
				}
			}
			if !coupled {
				h.peers[client] = nil
				log.Println("registered new client")
			}

		case client := <-h.unregister:
			_, ok := h.peers[client]
			if ok {
				delete(h.peers, client)
				close(client.send)
			}

		case message := <-h.send:
			log.Println("hub was requested to send a message")
			peer, ok := h.peers[message.Sender]
			if ok {
				select {
				case peer.send <- message.Text:
				default:
					close(peer.send)
					delete(h.peers, peer)
				}
			}
		}
	}
}
