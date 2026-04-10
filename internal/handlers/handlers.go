package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitpigeon/internal/domain"
	"gitpigeon/internal/service"

	"github.com/go-chi/chi/v5"
)

// SubscriptionHandler holds HTTP handlers for subscription endpoints.
type SubscriptionHandler struct {
	svc domain.SubscriptionService
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(svc domain.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// subscribeRequest is the JSON request body for POST /api/subscribe.
type subscribeRequest struct {
	Email string `json:"email"`
	Repo  string `json:"repo"`
}

// errorResponse is a standard error JSON response.
type errorResponse struct {
	Message string `json:"message"`
}

// HandleSubscribe handles POST /api/subscribe.
func (h *SubscriptionHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, errorResponse{Message: "invalid request body"})
		return
	}

	if req.Email == "" || req.Repo == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Message: "email and repo are required"})
		return
	}

	result, err := h.svc.Subscribe(r.Context(), req.Email, req.Repo)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// HandleConfirm handles GET /api/confirm/{token}.
// Returns a redirect on success or failure.
func (h *SubscriptionHandler) HandleConfirm(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Redirect(w, r, "/?status=error&message=missing+token", http.StatusFound)
		return
	}

	err := h.svc.Confirm(r.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrTokenNotFound) {
			http.Redirect(w, r, "/?status=error&message=invalid+or+expired+token", http.StatusFound)
			return
		}
		if errors.Is(err, service.ErrAlreadyConfirmed) {
			http.Redirect(w, r, "/?status=info&message=subscription+already+confirmed", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/?status=error&message=internal+error", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/?status=success&message=subscription+confirmed", http.StatusFound)
}

// HandleUnsubscribe handles GET /api/unsubscribe/{token}.
// Returns a redirect on success or failure.
func (h *SubscriptionHandler) HandleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Redirect(w, r, "/?status=error&message=missing+token", http.StatusFound)
		return
	}

	err := h.svc.Unsubscribe(r.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrTokenNotFound) {
			http.Redirect(w, r, "/?status=error&message=invalid+or+expired+token", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/?status=error&message=internal+error", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/?status=success&message=unsubscribed+successfully", http.StatusFound)
}

// HandleGetSubscriptions handles GET /api/subscriptions?email={email}.
func (h *SubscriptionHandler) HandleGetSubscriptions(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		respondJSON(w, http.StatusBadRequest, errorResponse{Message: "email query parameter is required"})
		return
	}

	subs, err := h.svc.GetSubscriptions(r.Context(), email)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, subs)
}

// handleServiceError maps service-layer errors to HTTP responses.
func (h *SubscriptionHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidRepoFormat):
		respondJSON(w, http.StatusBadRequest, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrInvalidEmail):
		respondJSON(w, http.StatusBadRequest, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrRepoNotFound):
		respondJSON(w, http.StatusNotFound, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrRateLimitExceeded):
		respondJSON(w, http.StatusTooManyRequests, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrSubscriptionExists):
		respondJSON(w, http.StatusConflict, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrTokenNotFound):
		respondJSON(w, http.StatusNotFound, errorResponse{Message: err.Error()})
	default:
		respondJSON(w, http.StatusInternalServerError, errorResponse{Message: "internal server error"})
	}
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
	}
}
