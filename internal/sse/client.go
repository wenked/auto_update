package sse

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Client struct {
	Id          string
	writer      http.ResponseWriter
	MessageChan chan string
}

func (c *Client) RunSSE() {
	c.writer.Header().Set("Content-Type", "text/event-stream")
	c.writer.Header().Set("Cache-Control", "no-cache")
	c.writer.Header().Set("Connection", "keep-alive")
	c.writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	flusher := c.writer.(http.Flusher)
	// initial connection

	_, err := c.writer.Write([]byte("data: connected\n\n"))

	if err != nil {
		log.Println("Error writing to client: ", err)
		return
	}

	flusher.Flush()

	for {
		select {
		case msg := <-c.MessageChan:
			fmt.Println("Sending message to client: ", msg)

			_, err := c.writer.Write([]byte("event: update\n" + "data: " + msg + "\n\n"))

			if err != nil {
				log.Println("Error writing to client: ", err)
				return
			}

			flusher.Flush()
		}
	}

}

func NewClient(w http.ResponseWriter) *Client {
	return &Client{
		Id:          uuid.NewString(),
		writer:      w,
		MessageChan: make(chan string),
	}
}
