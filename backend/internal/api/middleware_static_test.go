package api

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newRequestWithNonceContext(t *testing.T, method, path, nonce string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), NonceContextKey, nonce)
	return req.WithContext(ctx)
}

func writeStaticFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	index := `<html><body><script nonce="__CSP_NONCE__">window.ready=true;</script></body></html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(index), 0o644); err != nil {
		t.Fatalf("failed to write index.html: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('ok');"), 0o644); err != nil {
		t.Fatalf("failed to write app.js: %v", err)
	}

	return dir
}

func TestGenerateNonce(t *testing.T) {
	t.Parallel()

	nonceA, err := generateNonce()
	if err != nil {
		t.Fatalf("generateNonce() returned error: %v", err)
	}
	nonceB, err := generateNonce()
	if err != nil {
		t.Fatalf("generateNonce() returned error: %v", err)
	}

	decoded, err := base64.URLEncoding.DecodeString(nonceA)
	if err != nil {
		t.Fatalf("nonce is not valid URL-safe base64: %v", err)
	}
	if len(decoded) != 16 {
		t.Fatalf("decoded nonce length = %d, want 16", len(decoded))
	}
	if nonceA == nonceB {
		t.Fatalf("expected two generated nonces to differ, both were %q", nonceA)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	t.Parallel()

	var gotNonce string
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		nonce, ok := r.Context().Value(NonceContextKey).(string)
		if !ok || nonce == "" {
			t.Fatalf("missing nonce in request context")
		}
		gotNonce = nonce
		w.WriteHeader(http.StatusNoContent)
	})

	handler := securityHeadersMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Fatalf("expected next handler to be called")
	}

	csp := rr.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatalf("missing Content-Security-Policy header")
	}
	if !strings.Contains(csp, "'nonce-"+gotNonce+"'") {
		t.Fatalf("CSP does not contain request nonce %q: %q", gotNonce, csp)
	}

	if got := rr.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want %q", got, "nosniff")
	}
	if got := rr.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q, want %q", got, "DENY")
	}
	if got := rr.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Fatalf("Referrer-Policy = %q, want %q", got, "strict-origin-when-cross-origin")
	}
	if got := rr.Header().Get("Permissions-Policy"); got != "geolocation=(), microphone=(), camera=()" {
		t.Fatalf("Permissions-Policy = %q, want %q", got, "geolocation=(), microphone=(), camera=()")
	}
}

func TestCORSMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("options short-circuits", func(t *testing.T) {
		t.Parallel()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusNoContent)
		})

		handler := corsMiddleware(next)
		req := httptest.NewRequest(http.MethodOptions, "/api/inbox/abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if nextCalled {
			t.Fatalf("next handler should not be called for OPTIONS")
		}
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "*")
		}
	})

	t.Run("non-options calls next", func(t *testing.T) {
		t.Parallel()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusNoContent)
		})

		handler := corsMiddleware(next)
		req := httptest.NewRequest(http.MethodGet, "/api/inbox/abc", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
		}
		if !nextCalled {
			t.Fatalf("expected next handler to be called")
		}
		if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "*")
		}
		if got := rr.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, DELETE, OPTIONS" {
			t.Fatalf("Access-Control-Allow-Methods = %q, want %q", got, "GET, POST, DELETE, OPTIONS")
		}
		if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
			t.Fatalf("missing Access-Control-Allow-Headers header")
		}
	})
}

func TestResponseWriter(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	rw := newResponseWriter(rr)

	if rw.statusCode != http.StatusOK {
		t.Fatalf("initial statusCode = %d, want %d", rw.statusCode, http.StatusOK)
	}

	rw.WriteHeader(http.StatusTeapot)

	if rw.statusCode != http.StatusTeapot {
		t.Fatalf("captured statusCode = %d, want %d", rw.statusCode, http.StatusTeapot)
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("response recorder status = %d, want %d", rr.Code, http.StatusTeapot)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("captures default status 200", func(t *testing.T) {
		t.Parallel()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			_, _ = w.Write([]byte("ok"))
		})

		handler := loggingMiddleware(next)
		req := httptest.NewRequest(http.MethodGet, "/api/inbox/0123456789abcdef0123456789abcdef01234567", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if !nextCalled {
			t.Fatalf("expected next handler to be called")
		}
	})

	t.Run("captures explicit status", func(t *testing.T) {
		t.Parallel()

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusCreated)
		})

		handler := loggingMiddleware(next)
		req := httptest.NewRequest(http.MethodDelete, "/api/inbox/resource", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
		}
		if !nextCalled {
			t.Fatalf("expected next handler to be called")
		}
	})
}

func TestServeIndexWithNonce(t *testing.T) {
	t.Parallel()

	t.Run("missing nonce in context returns 500", func(t *testing.T) {
		t.Parallel()

		dir := writeStaticFixture(t)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		serveIndexWithNonce(rr, req, dir)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("missing index file returns 500", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		req := newRequestWithNonceContext(t, http.MethodGet, "/", "nonce-value")
		rr := httptest.NewRecorder()

		serveIndexWithNonce(rr, req, dir)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
	})

	t.Run("nonce placeholder replaced", func(t *testing.T) {
		t.Parallel()

		dir := writeStaticFixture(t)
		req := newRequestWithNonceContext(t, http.MethodGet, "/", "nonce-123")
		rr := httptest.NewRecorder()

		serveIndexWithNonce(rr, req, dir)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if got := rr.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want %q", got, "text/html; charset=utf-8")
		}
		if strings.Contains(rr.Body.String(), "__CSP_NONCE__") {
			t.Fatalf("response still contains nonce placeholder: %q", rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "nonce-123") {
			t.Fatalf("response body does not include nonce value: %q", rr.Body.String())
		}
	})
}

func TestServeStatic(t *testing.T) {
	t.Parallel()

	dir := writeStaticFixture(t)
	handler := serveStatic(dir)

	t.Run("root path serves index with nonce replacement", func(t *testing.T) {
		t.Parallel()

		req := newRequestWithNonceContext(t, http.MethodGet, "/", "nonce-root")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if !strings.Contains(rr.Body.String(), "nonce-root") {
			t.Fatalf("root response missing nonce replacement: %q", rr.Body.String())
		}
	})

	t.Run("index path serves index with nonce replacement", func(t *testing.T) {
		t.Parallel()

		req := newRequestWithNonceContext(t, http.MethodGet, "/index.html", "nonce-index")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if !strings.Contains(rr.Body.String(), "nonce-index") {
			t.Fatalf("index response missing nonce replacement: %q", rr.Body.String())
		}
	})

	t.Run("existing asset served directly", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if body := rr.Body.String(); body != "console.log('ok');" {
			t.Fatalf("asset body = %q, want %q", body, "console.log('ok');")
		}
	})

	t.Run("missing asset falls back to index", func(t *testing.T) {
		t.Parallel()

		req := newRequestWithNonceContext(t, http.MethodGet, "/missing.js", "nonce-fallback")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if !strings.Contains(rr.Body.String(), "nonce-fallback") {
			t.Fatalf("fallback response missing nonce replacement: %q", rr.Body.String())
		}
	})
}
