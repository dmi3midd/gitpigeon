package server

import (
	"encoding/json"
	"net/http"

	"gitpigeon/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", s.healthHandler)

	r.Route("/api", func(r chi.Router) {
		r.Get("/confirm/{token}", s.handler.HandleConfirm)
		r.Get("/unsubscribe/{token}", s.handler.HandleUnsubscribe)

		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKeyAuth(s.apiKey))

			r.Post("/subscribe", s.handler.HandleSubscribe)
			r.Get("/subscriptions", s.handler.HandleGetSubscriptions)
		})
	})

	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(jsonResp)
}
