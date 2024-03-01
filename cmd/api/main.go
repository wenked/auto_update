package main

import (
	logger "auto-update/config"
	"auto-update/internal/queue"
	"auto-update/internal/server"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var osSignal chan os.Signal

func main() {

	osSignal = make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	queue := queue.NewUpdateQueue()

	server := server.NewServer(queue)
	logger.InitLogger()

	slog.Info("starting api on port", os.Getenv("PORT"), "...")
	// Start the server concurrently
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Unexpected server error because of: %v\n", err)
		}
	}()

	go queue.Work()

	<-osSignal

	fmt.Println("Terminating server")
	server.Shutdown(context.Background())

	fmt.Println("Terminating update queue")

	for queue.Size() > 0 {
		time.Sleep(time.Millisecond * 500)
	}

	fmt.Println("Complete terminating application")

}
