package server

import (
	"fmt"
	"net/http"
	"time"

	"gitpigeon/internal/config"
	"gitpigeon/internal/database"
	githubapi "gitpigeon/internal/github"
	"gitpigeon/internal/handlers"
	"gitpigeon/internal/notifier"
	"gitpigeon/internal/repositories"
	"gitpigeon/internal/service"
)

type Server struct {
	port   int
	apiKey string

	db      database.Service
	handler *handlers.SubscriptionHandler
}

func NewServer(cfg *config.Config) (*http.Server, error) {
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	repoRepo := repositories.NewRepositoryRepo(db.DB())
	subRepo := repositories.NewSubscriptionRepo(db.DB())

	// Initialize GitHub client
	ghClient := githubapi.NewClient(cfg.GitHubToken)

	// Initialize email notifier
	emailNotifier, err := notifier.NewEmailNotifier(&cfg.SMTP)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize notifier: %w", err)
	}

	// Initialize service layer
	subService := service.NewSubscriptionService(
		subRepo,
		repoRepo,
		ghClient,
		emailNotifier,
		cfg.AppBaseURL,
	)

	// Initialize handlers
	subHandler := handlers.NewSubscriptionHandler(subService)

	s := &Server{
		port:    cfg.AppPort,
		apiKey:  cfg.ApiKey,
		db:      db,
		handler: subHandler,
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
