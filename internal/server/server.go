package server

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"gitpigeon/internal/config"
	"gitpigeon/internal/database"
)

type Server struct {
	port int

	db database.Service
}

func NewServer(cfg *config.Config) *http.Server {
	port := cfg.AppPort
	NewServer := &Server{
		port: port,

		db: database.New(cfg),
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
