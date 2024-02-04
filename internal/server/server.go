package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"auto-update/internal/database"
	"auto-update/internal/queue"
	"auto-update/internal/sse"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port  int
	db    database.Service
	queue *queue.UpdateQueue
	hub   *sse.Hub
}

func NewServer(queue *queue.UpdateQueue) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	hub := sse.GetHub()
	go hub.Run()

	NewServer := &Server{
		port:  port,
		db:    database.GetService(),
		queue: queue,
		hub:   hub,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
