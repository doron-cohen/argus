package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/doron-cohen/argus/backend/internal/config"
	"github.com/doron-cohen/argus/backend/internal/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	stop, err := server.Start(cfg)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	stop()
}
