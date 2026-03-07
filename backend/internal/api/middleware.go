package api

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/metrics"
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
				// Track rate limit hits
				metrics.RateLimitHitsTotal.WithLabelValues(config.KeyPrefix).Inc()
				writeError(w, ErrCodeRateLimitExceeded, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type contextKey string

const NonceContextKey contextKey = "csp-nonce"

func generateNonce() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := generateNonce()
		if err != nil {
			log.Printf("Failed to generate CSP nonce: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		csp := fmt.Sprintf(
			"default-src 'self'; "+
				"script-src 'self' 'nonce-%s' 'strict-dynamic'; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src 'self' https://fonts.gstatic.com; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self'; "+
				"object-src 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'; "+
				"frame-ancestors 'none'",
			nonce,
		)
		w.Header().Set("Content-Security-Policy", csp)

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		ctx := context.WithValue(r.Context(), NonceContextKey, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Public-Key, X-Signature, X-Timestamp, X-Nonce")
		// TODO: restrict to actual domain in production

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Normalize endpoint path (replace dynamic segments)
		endpoint := normalizeEndpoint(r.URL.Path)

		// Track in-flight requests
		metrics.HTTPRequestsInFlight.WithLabelValues(endpoint).Inc()
		defer metrics.HTTPRequestsInFlight.WithLabelValues(endpoint).Dec()

		// Wrap response writer to capture status code
		wrapped := newResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusStr := strconv.Itoa(wrapped.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, statusStr).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

// normalizeEndpoint converts dynamic paths to templates for consistent metrics
func normalizeEndpoint(path string) string {
	// Replace UUID-like segments and hash-like segments with placeholders
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		// Check if it looks like an address (40 hex chars) or email ID (UUID/hash)
		if len(part) >= 32 || (len(part) > 8 && isHexOrAlphanumeric(part)) {
			parts[i] = "{id}"
		}
	}
	normalized := strings.Join(parts, "/")
	if normalized == "" {
		return "/"
	}
	return normalized
}

func isHexOrAlphanumeric(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func signatureAuthMiddleware(s store.EmailStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pubKeyHex := r.Header.Get("X-Public-Key")
			sigHex := r.Header.Get("X-Signature")
			tsStr := r.Header.Get("X-Timestamp")
			nonce := r.Header.Get("X-Nonce")

			if pubKeyHex == "" || sigHex == "" || tsStr == "" || nonce == "" {
				metrics.AuthFailuresTotal.WithLabelValues("missing_headers").Inc()
				writeError(w, ErrCodeUnauthorized, "Missing authentication headers", http.StatusUnauthorized)
				return
			}

			if _, err := uuid.Parse(nonce); err != nil {
				metrics.AuthFailuresTotal.WithLabelValues("invalid_nonce").Inc()
				writeError(w, ErrCodeUnauthorized, "Invalid nonce format", http.StatusUnauthorized)
				return
			}

			ts, err := strconv.ParseInt(tsStr, 10, 64)
			if err != nil {
				metrics.AuthFailuresTotal.WithLabelValues("invalid_timestamp").Inc()
				writeError(w, ErrCodeUnauthorized, "Invalid timestamp format", http.StatusUnauthorized)
				return
			}

			clientTime := time.Unix(ts, 0)
			timeDiff := time.Since(clientTime)
			if timeDiff > 5*time.Minute || timeDiff < -5*time.Minute {
				metrics.AuthFailuresTotal.WithLabelValues("expired_timestamp").Inc()
				writeError(w, ErrCodeUnauthorized, "Request expired or invalid timestamp", http.StatusUnauthorized)
				return
			}

			pubKeyBytes, err := hex.DecodeString(pubKeyHex)
			if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
				metrics.AuthFailuresTotal.WithLabelValues("invalid_pubkey").Inc()
				writeError(w, ErrCodeUnauthorized, "Invalid public key format", http.StatusUnauthorized)
				return
			}

			sigBytes, err := hex.DecodeString(sigHex)
			if err != nil || len(sigBytes) != ed25519.SignatureSize {
				metrics.AuthFailuresTotal.WithLabelValues("invalid_signature_format").Inc()
				writeError(w, ErrCodeUnauthorized, "Invalid signature format", http.StatusUnauthorized)
				return
			}

			hash := sha256.Sum256(pubKeyBytes)
			derivedAddress := hex.EncodeToString(hash[:])[:40]

			address := r.PathValue("address")
			if address == "" {
				metrics.AuthFailuresTotal.WithLabelValues("missing_address").Inc()
				writeError(w, ErrCodeUnauthorized, "Missing address parameter", http.StatusBadRequest)
				return
			}
			if address != derivedAddress {
				metrics.AuthFailuresTotal.WithLabelValues("address_mismatch").Inc()
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

			log.Printf("[VERIFY DEBUG] method=%s path=%s ts=%s bodyLen=%d bodyHash=%s nonce=%s payload=%s",
				r.Method, r.URL.Path, tsStr, len(bodyBytes), bodyHashHex, nonce, payload)

			if !ed25519.Verify(pubKeyBytes, []byte(payload), sigBytes) {
				metrics.AuthFailuresTotal.WithLabelValues("signature_verification_failed").Inc()
				writeError(w, ErrCodeUnauthorized, "Invalid cryptographic signature", http.StatusUnauthorized)
				return
			}

			unique, err := s.CheckAndStoreNonce(r.Context(), nonce, 5*time.Minute)
			if err != nil {
				log.Printf("Nonce check error: %v", err)
				writeError(w, ErrCodeInternalError, "Failed to verify nonce", http.StatusInternalServerError)
				return
			}
			if !unique {
				metrics.AuthFailuresTotal.WithLabelValues("nonce_reuse").Inc()
				writeError(w, ErrCodeUnauthorized, "Nonce already used", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func serveIndexWithNonce(w http.ResponseWriter, r *http.Request, staticDir string) {
	nonce, ok := r.Context().Value(NonceContextKey).(string)
	if !ok {
		log.Println("Missing nonce in request context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	filePath := staticDir + "/index.html"
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read index.html: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	html := strings.ReplaceAll(string(content), `__CSP_NONCE__`, nonce)
	w.Write([]byte(html))
}

func serveStatic(staticDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		staticFs := http.FileServer(http.Dir(staticDir))

		path := r.URL.Path
		if path == "/" || path == "" {
			path = "/index.html"
		}

		if path == "/index.html" {
			serveIndexWithNonce(w, r, staticDir)
			return
		}

		filePath := staticDir + path
		if _, err := os.Stat(filePath); err != nil {
			serveIndexWithNonce(w, r, staticDir)
			return
		}

		staticFs.ServeHTTP(w, r)
	}
}
