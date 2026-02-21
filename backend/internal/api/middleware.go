package api

import (
	"log"
	"net/http"
	"time"

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
