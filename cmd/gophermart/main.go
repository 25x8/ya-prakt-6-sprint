package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/config"
	"github.com/25x8/ya-prakt-6-sprint/internal/gophermart/server"
)

func main() {
	// Load configuration
	cfg := config.NewConfig()

	// Create and run server
	srv := server.NewServer(cfg)
	go func() {
		if err := srv.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
