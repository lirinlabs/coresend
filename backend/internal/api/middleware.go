package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/addr"
	"github.com/fn-jakubkarp/coresend/internal/store"
)

type RateLimitConfig struct {
	Limit     int
	Window    time.Duration
	KeyPrefix string
}

func rateLimitMiddleware(s store.EmailStore, config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			key := config.KeyPrefix + ":" + clientIP

			allowed, _, err := s.CheckRateLimit(r.Context(), key, config.Limit, config.Window)
			if err != nil {
				log.Printf("Rate limit check error: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				writeError(w, ErrCodeRateLimitExceeded, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Auth-Address")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addressHeader := r.Header.Get("X-Auth-Address")

		if addressHeader == "" {
			writeError(w, ErrCodeUnauthorized, "Missing X-Auth-Address header", http.StatusUnauthorized)
			return
		}

		if !addr.IsValid(addressHeader) {
			writeError(w, ErrCodeUnauthorized, "Invalid address format", http.StatusUnauthorized)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		var pathAddress string
		for i, part := range pathParts {
			if part == "inbox" && i+1 < len(pathParts) {
				pathAddress = pathParts[i+1]
				break
			}
		}

		if pathAddress == "" {
			writeError(w, ErrCodeUnauthorized, "Invalid request path", http.StatusUnauthorized)
			return
		}

		if !strings.EqualFold(pathAddress, addressHeader) {
			writeError(w, ErrCodeUnauthorized, "Address mismatch", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
