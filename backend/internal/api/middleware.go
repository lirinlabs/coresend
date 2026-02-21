package api

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/store"
	"github.com/google/uuid"
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Public-Key, X-Signature, X-Timestamp, X-Nonce")

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

func signatureAuthMiddleware(s store.EmailStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pubKeyHex := r.Header.Get("X-Public-Key")
			sigHex := r.Header.Get("X-Signature")
			tsStr := r.Header.Get("X-Timestamp")
			nonce := r.Header.Get("X-Nonce")

			if pubKeyHex == "" || sigHex == "" || tsStr == "" || nonce == "" {
				writeError(w, ErrCodeUnauthorized, "Missing authentication headers", http.StatusUnauthorized)
				return
			}

			if _, err := uuid.Parse(nonce); err != nil {
				writeError(w, ErrCodeUnauthorized, "Invalid nonce format", http.StatusUnauthorized)
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

			unique, err := s.CheckAndStoreNonce(r.Context(), nonce, 5*time.Minute)
			if err != nil {
				log.Printf("Nonce check error: %v", err)
				writeError(w, ErrCodeInternalError, "Failed to verify nonce", http.StatusInternalServerError)
				return
			}
			if !unique {
				writeError(w, ErrCodeUnauthorized, "Nonce already used", http.StatusUnauthorized)
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

			address := r.PathValue("address")
			if address != derivedAddress {
				writeError(w, ErrCodeUnauthorized, "Access denied: address does not match public key", http.StatusForbidden)
				return
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				writeError(w, ErrCodeInternalError, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			bodyHash := sha256.Sum256(bodyBytes)
			bodyHashHex := hex.EncodeToString(bodyHash[:])

			payload := fmt.Sprintf("%s:%s:%s:%s:%s", r.Method, r.URL.Path, tsStr, bodyHashHex, nonce)

			if !ed25519.Verify(pubKeyBytes, []byte(payload), sigBytes) {
				writeError(w, ErrCodeUnauthorized, "Invalid cryptographic signature", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
