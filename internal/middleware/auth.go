package middleware

import (
	"net/http"
)

// APIKeyAuth returns a middleware that validates the X-API-Key header.
func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" || key != apiKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"message":"unauthorized: missing or invalid API key"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
