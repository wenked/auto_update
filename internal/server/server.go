package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"auto-update/internal/database"
	"auto-update/internal/queue"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
	db   database.Service
	queue *queue.UpdateQueue
}

func NewServer(queue *queue.UpdateQueue) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	
	NewServer := &Server{
		port: port,
		db:   database.New(),
		queue: queue,
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
