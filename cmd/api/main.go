package main

import (
	logger "auto-update/config"
	"auto-update/internal/server"
	"fmt"
	"log/slog"
)

func main() {

	server := server.NewServer()
	logger.InitLogger()

    slog.Info("starting api on port 8080")
	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
