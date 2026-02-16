package api

import (
	"net/http"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/store"
	httpSwagger "github.com/swaggo/http-swagger"
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

	mux.HandleFunc("/api/identity/generate", wrap(handler.handleGenerateMnemonic, loggingMiddleware, corsMiddleware, rateLimitMiddleware(s, generateRateLimit)))
	mux.HandleFunc("/api/identity/derive", wrap(handler.handleDeriveAddress, loggingMiddleware, corsMiddleware))
	mux.HandleFunc("/api/identity/validate/", wrap(handler.handleValidateAddress, loggingMiddleware, corsMiddleware))
	mux.HandleFunc("/api/inbox/", wrap(handler.handleGetInbox, loggingMiddleware, corsMiddleware, rateLimitMiddleware(s, inboxRateLimit), authMiddleware))
	mux.HandleFunc("/api/inbox", wrap(handler.handleClearInbox, loggingMiddleware, corsMiddleware, rateLimitMiddleware(s, deleteRateLimit), authMiddleware))
	mux.HandleFunc("/api/health", wrap(handler.handleHealth, loggingMiddleware, corsMiddleware))
	mux.HandleFunc("/docs/", httpSwagger.WrapHandler)

	return mux
}

func wrap(handler http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.HandlerFunc {
	wrapped := http.Handler(handler)
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
	}
}
