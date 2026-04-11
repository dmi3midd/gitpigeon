package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"gitpigeon/internal/config"
	"gitpigeon/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	httpServer, app, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Create a root context that is cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the release scanner in a background goroutine
	go app.Scanner.Start(ctx)

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("server: listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	stop() // Allow Ctrl+C to force shutdown

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// Give the server 5 seconds to finish in-flight requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown with error: %v", err)
	}

	log.Println("graceful shutdown complete")
}
