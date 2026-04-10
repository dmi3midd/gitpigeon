package server

import (
	"fmt"
	"net/http"
	"time"

	"gitpigeon/internal/config"
	"gitpigeon/internal/database"
)

type Server struct {
	port int

	db database.Service
}

func NewServer(cfg *config.Config) (*http.Server, error) {
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	s := &Server{
		port: cfg.AppPort,
		db:   db,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}
