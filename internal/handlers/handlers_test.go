package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitpigeon/internal/domain"
	"gitpigeon/internal/mocks"
	"gitpigeon/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRouter creates a chi router with the handler registered for testing.
func newTestRouter(h *SubscriptionHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/subscribe", h.HandleSubscribe)
	r.Get("/api/confirm/{token}", h.HandleConfirm)
	r.Get("/api/unsubscribe/{token}", h.HandleUnsubscribe)
	r.Get("/api/subscriptions", h.HandleGetSubscriptions)
	return r
}

// --- HandleSubscribe tests ---

func TestHandleSubscribe_Success(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		SubscribeFn: func(ctx context.Context, email, repo string) (*domain.SubscribeResult, error) {
			return &domain.SubscribeResult{Message: "Confirmation email sent."}, nil
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	body := `{"email":"user@example.com","repo":"golang/go"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.SubscribeResult
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Message, "Confirmation email sent")
}

func TestHandleSubscribe_InvalidBody(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		SubscribeFn: func(ctx context.Context, email, repo string) (*domain.SubscribeResult, error) {
			return nil, nil
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleSubscribe_MissingFields(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		SubscribeFn: func(ctx context.Context, email, repo string) (*domain.SubscribeResult, error) {
			return nil, nil
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	tests := []struct {
		name string
		body string
	}{
		{"missing email", `{"repo":"golang/go"}`},
		{"missing repo", `{"email":"user@example.com"}`},
		{"empty both", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestHandleSubscribe_RepoNotFound(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		SubscribeFn: func(ctx context.Context, email, repo string) (*domain.SubscribeResult, error) {
			return nil, service.ErrRepoNotFound
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	body := `{"email":"user@example.com","repo":"nonexistent/repo"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleSubscribe_RateLimit(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		SubscribeFn: func(ctx context.Context, email, repo string) (*domain.SubscribeResult, error) {
			return nil, service.ErrRateLimitExceeded
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	body := `{"email":"user@example.com","repo":"golang/go"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestHandleSubscribe_Conflict(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		SubscribeFn: func(ctx context.Context, email, repo string) (*domain.SubscribeResult, error) {
			return nil, service.ErrSubscriptionExists
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	body := `{"email":"user@example.com","repo":"golang/go"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- HandleConfirm tests ---

func TestHandleConfirm_Success_Redirects(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		ConfirmFn: func(ctx context.Context, token string) error {
			return nil
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/valid-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "status=success")
}

func TestHandleConfirm_InvalidToken_Redirects(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		ConfirmFn: func(ctx context.Context, token string) error {
			return service.ErrTokenNotFound
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/bad-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "status=error")
}

func TestHandleConfirm_AlreadyConfirmed_Redirects(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		ConfirmFn: func(ctx context.Context, token string) error {
			return service.ErrAlreadyConfirmed
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/some-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "status=info")
}

// --- HandleUnsubscribe tests ---

func TestHandleUnsubscribe_Success_Redirects(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		UnsubscribeFn: func(ctx context.Context, token string) error {
			return nil
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/valid-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "status=success")
}

func TestHandleUnsubscribe_InvalidToken_Redirects(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		UnsubscribeFn: func(ctx context.Context, token string) error {
			return service.ErrTokenNotFound
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/bad-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "status=error")
}

// --- HandleGetSubscriptions tests ---

func TestHandleGetSubscriptions_Success(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		GetSubscriptionsFn: func(ctx context.Context, email string) ([]domain.SubscriptionInfo, error) {
			return []domain.SubscriptionInfo{
				{ID: 1, Repo: "golang/go", Email: "user@example.com", CreatedAt: "2026-04-12T00:00:00Z"},
			}, nil
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions?email=user@example.com", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var subs []domain.SubscriptionInfo
	err := json.Unmarshal(w.Body.Bytes(), &subs)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Equal(t, "golang/go", subs[0].Repo)
}

func TestHandleGetSubscriptions_MissingEmail(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		GetSubscriptionsFn: func(ctx context.Context, email string) ([]domain.SubscriptionInfo, error) {
			return nil, nil
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetSubscriptions_InvalidEmail(t *testing.T) {
	mockSvc := &mocks.MockSubscriptionService{
		GetSubscriptionsFn: func(ctx context.Context, email string) ([]domain.SubscriptionInfo, error) {
			return nil, service.ErrInvalidEmail
		},
	}
	h := NewSubscriptionHandler(mockSvc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions?email=bad", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
