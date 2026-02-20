package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	testAddress := "b4ebe3e2200cbc90"

	tests := []struct {
		name           string
		authAddress    string
		path           string
		expectedStatus int
	}{
		{
			name:           "valid authorization",
			authAddress:    testAddress,
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing X-Auth-Address",
			authAddress:    "",
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid address format",
			authAddress:    "invalid",
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "path address mismatch",
			authAddress:    testAddress,
			path:           "/api/inbox/0000000000000000",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid path",
			authAddress:    testAddress,
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
			if tt.authAddress != "" {
				req.Header.Set("X-Auth-Address", tt.authAddress)
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
