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

	inboxLimit := RateLimitConfig{Limit: 60, Window: time.Minute, KeyPrefix: "inbox"}
	deleteLimit := RateLimitConfig{Limit: 30, Window: time.Minute, KeyPrefix: "delete"}

	mux.HandleFunc("POST /api/register/{address}", wrap(handler.handleRegister, loggingMiddleware, corsMiddleware, signatureAuthMiddleware))

	mux.HandleFunc("GET /api/inbox/{address}", wrap(handler.handleGetInbox, loggingMiddleware, corsMiddleware, signatureAuthMiddleware, rateLimitMiddleware(s, inboxLimit)))
	mux.HandleFunc("GET /api/inbox/{address}/{emailId}", wrap(handler.handleGetEmail, loggingMiddleware, corsMiddleware, signatureAuthMiddleware, rateLimitMiddleware(s, inboxLimit)))
	mux.HandleFunc("DELETE /api/inbox/{address}/{emailId}", wrap(handler.handleDeleteEmail, loggingMiddleware, corsMiddleware, signatureAuthMiddleware, rateLimitMiddleware(s, deleteLimit)))
	mux.HandleFunc("DELETE /api/inbox/{address}", wrap(handler.handleClearInbox, loggingMiddleware, corsMiddleware, signatureAuthMiddleware, rateLimitMiddleware(s, deleteLimit)))

	mux.HandleFunc("GET /api/health", wrap(handler.handleHealth, loggingMiddleware, corsMiddleware))
	mux.HandleFunc("GET /docs/", httpSwagger.WrapHandler)

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
