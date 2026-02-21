package api

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Public-Key, X-Signature, X-Timestamp")

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

func signatureAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pubKeyHex := r.Header.Get("X-Public-Key")
		sigHex := r.Header.Get("X-Signature")
		tsStr := r.Header.Get("X-Timestamp")

		if pubKeyHex == "" || sigHex == "" || tsStr == "" {
			writeError(w, ErrCodeUnauthorized, "Missing authentication headers", http.StatusUnauthorized)
			return
		}

		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			writeError(w, ErrCodeUnauthorized, "Invalid timestamp format", http.StatusUnauthorized)
			return
		}

		clientTime := time.Unix(ts, 0)
		timeDiff := time.Since(clientTime)
		if timeDiff > 5*time.Minute || timeDiff < -5*time.Minute {
			writeError(w, ErrCodeUnauthorized, "Request expired or invalid timestamp", http.StatusUnauthorized)
			return
		}

		pubKeyBytes, err := hex.DecodeString(pubKeyHex)
		if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
			writeError(w, ErrCodeUnauthorized, "Invalid public key format", http.StatusUnauthorized)
			return
		}

		sigBytes, err := hex.DecodeString(sigHex)
		if err != nil || len(sigBytes) != ed25519.SignatureSize {
			writeError(w, ErrCodeUnauthorized, "Invalid signature format", http.StatusUnauthorized)
			return
		}

		hash := sha256.Sum256(pubKeyBytes)
		derivedAddress := hex.EncodeToString(hash[:])[:40]

		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 || parts[2] != derivedAddress {
			writeError(w, ErrCodeUnauthorized, "Access denied: address does not match public key", http.StatusForbidden)
			return
		}

		payload := fmt.Sprintf("%s:%s:%s", r.Method, r.URL.Path, tsStr)

		if !ed25519.Verify(pubKeyBytes, []byte(payload), sigBytes) {
			writeError(w, ErrCodeUnauthorized, "Invalid cryptographic signature", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
