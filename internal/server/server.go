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
	"auto-update/internal/sshclient"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port      int
	db        database.Service
	queue     *queue.UpdateQueue
	hub       *sse.Hub
	sshclient *sshclient.SshClientService
}

func NewServer(queue *queue.UpdateQueue) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	hub := sse.GetHub()

	// spawn 10 workers
	// each worker will process 1 update at a time
	//for i := 0; i < 10; i++ {
	go hub.Run()
	//}

	NewServer := &Server{
		port:      port,
		db:        database.GetService(),
		queue:     queue,
		hub:       hub,
		sshclient: sshclient.NewSshClientService(),
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
