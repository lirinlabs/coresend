package api

import (
	"crypto/ed25519"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/identity"
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Auth-Address, X-Auth-Timestamp, X-Auth-Pubkey, X-Auth-Signature")

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
		timestampHeader := r.Header.Get("X-Auth-Timestamp")
		pubkeyHeader := r.Header.Get("X-Auth-Pubkey")
		signatureHeader := r.Header.Get("X-Auth-Signature")

		if addressHeader == "" || timestampHeader == "" || pubkeyHeader == "" || signatureHeader == "" {
			writeError(w, ErrCodeUnauthorized, "Missing authentication headers", http.StatusUnauthorized)
			return
		}

		if !identity.IsValidAddress(addressHeader) {
			writeError(w, ErrCodeUnauthorized, "Invalid address format", http.StatusUnauthorized)
			return
		}

		timestamp, err := strconv.ParseInt(timestampHeader, 10, 64)
		if err != nil {
			writeError(w, ErrCodeUnauthorized, "Invalid timestamp format", http.StatusUnauthorized)
			return
		}

		now := time.Now().UnixMilli()
		if timestamp < now-60000 || timestamp > now+60000 {
			writeError(w, ErrCodeUnauthorized, "Timestamp expired or future", http.StatusUnauthorized)
			return
		}

		pubkey, err := hex.DecodeString(pubkeyHeader)
		if err != nil || len(pubkey) != 32 {
			writeError(w, ErrCodeUnauthorized, "Invalid public key format", http.StatusUnauthorized)
			return
		}

		signature, err := hex.DecodeString(signatureHeader)
		if err != nil || len(signature) != 64 {
			writeError(w, ErrCodeUnauthorized, "Invalid signature format", http.StatusUnauthorized)
			return
		}

		derivedAddress := identity.AddressFromPublicKey(pubkey)
		if derivedAddress != strings.ToLower(addressHeader) {
			log.Printf("Auth failed: derived=%s, header=%s", derivedAddress, strings.ToLower(addressHeader))
			writeError(w, ErrCodeUnauthorized, "Address does not match public key", http.StatusUnauthorized)
			return
		}

		message := identity.CreateMessageToSign(addressHeader, timestamp)
		if !ed25519.Verify(pubkey, []byte(message), signature) {
			log.Printf("Signature verification failed for address=%s", addressHeader)
			writeError(w, ErrCodeUnauthorized, "Invalid signature", http.StatusUnauthorized)
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

		if strings.ToLower(pathAddress) != strings.ToLower(addressHeader) {
			writeError(w, ErrCodeUnauthorized, "Address mismatch", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
