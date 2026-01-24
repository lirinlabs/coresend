package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fn-jakubkarp/coresend/internal/identity"
)

func TestAuthMiddleware(t *testing.T) {
	testMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	testAddress := identity.AddressFromMnemonic(testMnemonic)
	invalidMnemonic := "invalid mnemonic phrase here"

	tests := []struct {
		name           string
		authHeader     string
		path           string
		expectedStatus int
	}{
		{
			name:           "valid authorization",
			authHeader:     "Bearer " + testMnemonic,
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no authorization header",
			authHeader:     "",
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization format",
			authHeader:     "InvalidFormat " + testMnemonic,
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty mnemonic",
			authHeader:     "Bearer ",
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid mnemonic",
			authHeader:     "Bearer " + invalidMnemonic,
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong address",
			authHeader:     "Bearer " + testMnemonic,
			path:           "/api/inbox/0000000000000000",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid path",
			authHeader:     "Bearer " + testMnemonic,
			path:           "/api/invalid",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			handler := authMiddleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestWrapMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	corsApplied := false
	testCorsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			corsApplied = true
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}

	wrapped := wrap(handler, testCorsMiddleware)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	wrapped(w, req)

	if !corsApplied {
		t.Error("CORS middleware was not applied")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
