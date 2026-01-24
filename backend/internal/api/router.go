package api

import (
	"net/http"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/store"
)

func NewRouter(s store.EmailStore, domain string) http.Handler {
	handler := NewAPIHandler(s, domain)
	mux := http.NewServeMux()

	generateRateLimit := RateLimitConfig{
		Limit:     10,
		Window:    time.Minute,
		KeyPrefix: "generate",
	}

	inboxRateLimit := RateLimitConfig{
		Limit:     60,
		Window:    time.Minute,
		KeyPrefix: "inbox",
	}

	deleteRateLimit := RateLimitConfig{
		Limit:     30,
		Window:    time.Minute,
		KeyPrefix: "delete",
	}

	mux.HandleFunc("/api/identity/generate", loggingMiddleware(corsMiddleware(rateLimitMiddleware(s, generateRateLimit)(http.HandlerFunc(handler.handleGenerateMnemonic)))).ServeHTTP)
	mux.HandleFunc("/api/identity/derive", loggingMiddleware(corsMiddleware(http.HandlerFunc(handler.handleDeriveAddress))).ServeHTTP)
	mux.HandleFunc("/api/identity/validate/", loggingMiddleware(corsMiddleware(http.HandlerFunc(handler.handleValidateAddress))).ServeHTTP)
	mux.HandleFunc("/api/inbox/", loggingMiddleware(corsMiddleware(rateLimitMiddleware(s, inboxRateLimit)(http.HandlerFunc(handler.handleGetInbox)))).ServeHTTP)
	mux.HandleFunc("/api/inbox", loggingMiddleware(corsMiddleware(rateLimitMiddleware(s, deleteRateLimit)(http.HandlerFunc(handler.handleClearInbox)))).ServeHTTP)
	mux.HandleFunc("/api/health", loggingMiddleware(corsMiddleware(http.HandlerFunc(handler.handleHealth))).ServeHTTP)

	return mux
}
