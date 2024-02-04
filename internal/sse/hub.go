package sse

import "fmt"

type Hub struct {
	clients   map[string]*Client
	Broadcast chan string
	AddClient chan *Client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.AddClient:
			h.clients[client.Id] = client
		case msg := <-h.Broadcast:
			fmt.Println("Broadcasting message to clients: ", msg)
			for _, client := range h.clients {
				client.MessageChan <- msg
			}
		}
	}
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[string]*Client),
		Broadcast: make(chan string),
		AddClient: make(chan *Client),
	}
}

var hub = NewHub()

func GetHub() *Hub {
	return hub
}
