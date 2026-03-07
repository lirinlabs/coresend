package api

import (
	"bytes"
	"context"
	"crypto/ed25519"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/store"
)

func newSignedRouteRequest(t *testing.T, method, pathTemplate string, body []byte, at time.Time) (*http.Request, string) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 keypair: %v", err)
	}

	hash := sha256.Sum256(publicKey)
	address := hex.EncodeToString(hash[:])[:40]
	path := strings.Replace(pathTemplate, "{address}", address, 1)

	nonce := "6f32f622-e2be-47f3-bdb4-80c77a7b4f48"
	timestamp := strconv.FormatInt(at.Unix(), 10)
	bodyHash := sha256.Sum256(body)
	bodyHashHex := hex.EncodeToString(bodyHash[:])
	payload := fmt.Sprintf("%s:%s:%s:%s:%s", method, path, timestamp, bodyHashHex, nonce)
	signature := ed25519.Sign(privateKey, []byte(payload))

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("X-Public-Key", hex.EncodeToString(publicKey))
	req.Header.Set("X-Signature", hex.EncodeToString(signature))
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)

	return req, address
}

func TestWrap_Order(t *testing.T) {
	t.Parallel()

	callOrder := make([]string, 0, 7)

	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callOrder = append(callOrder, name+"_before")
				next.ServeHTTP(w, r)
				callOrder = append(callOrder, name+"_after")
			})
		}
	}

	handler := wrap(
		func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "handler")
			w.WriteHeader(http.StatusNoContent)
		},
		mw("first"),
		mw("second"),
		mw("third"),
	)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}

	wantOrder := []string{
		"first_before",
		"second_before",
		"third_before",
		"handler",
		"third_after",
		"second_after",
		"first_after",
	}
	if !slices.Equal(callOrder, wantOrder) {
		t.Fatalf("middleware execution order = %v, want %v", callOrder, wantOrder)
	}
}

func TestNewRouter_RegisterRoute_WithValidAuth(t *testing.T) {
	t.Parallel()

	fakeStore := &fakeEmailStore{}
	router := NewRouter(fakeStore, "coresend.dev", writeStaticFixture(t))

	req, address := newSignedRouteRequest(t, http.MethodPost, "/api/register/{address}", nil, time.Now())
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if fakeStore.registerCallCount != 1 {
		t.Fatalf("register call count = %d, want %d", fakeStore.registerCallCount, 1)
	}
	if fakeStore.lastRegisterAddress != address {
		t.Fatalf("register address = %q, want %q", fakeStore.lastRegisterAddress, address)
	}
	if fakeStore.lastRegisterDuration != 24*time.Hour {
		t.Fatalf("register ttl = %s, want %s", fakeStore.lastRegisterDuration, 24*time.Hour)
	}

	resp := decodeJSONResponse[RegisterResponse](t, rr)
	if !resp.Registered {
		t.Fatalf("registered = %v, want true", resp.Registered)
	}
	if resp.Address != address {
		t.Fatalf("response address = %q, want %q", resp.Address, address)
	}
}

func TestNewRouter_ProtectedRoutes_RejectMissingAuth(t *testing.T) {
	t.Parallel()

	fakeStore := &fakeEmailStore{}
	router := NewRouter(fakeStore, "coresend.dev", writeStaticFixture(t))

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "inbox route",
			method: http.MethodGet,
			path:   "/api/inbox/" + testValidAddress,
		},
		{
			name:   "delete route",
			method: http.MethodDelete,
			path:   "/api/inbox/" + testValidAddress + "/email-1",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
			}
			errResp := decodeErrorResponse(t, rr)
			if errResp.Error.Code != ErrCodeUnauthorized {
				t.Fatalf("error.code = %q, want %q", errResp.Error.Code, ErrCodeUnauthorized)
			}
		})
	}
}

func TestNewRouter_ProtectedRoutes_PassWithValidAuth(t *testing.T) {
	t.Parallel()

	mailTime := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	fakeStore := &fakeEmailStore{
		getEmailsFn: func(ctx context.Context, addressBox string) ([]store.Email, error) {
			return []store.Email{
				{
					ID:         "id-1",
					From:       "sender@example.com",
					To:         []string{addressBox + "@coresend.dev"},
					Subject:    "Subject",
					Body:       "Body",
					ReceivedAt: mailTime,
				},
			}, nil
		},
		deleteEmailFn: func(ctx context.Context, addressBox string, emailID string) error {
			return nil
		},
	}
	router := NewRouter(fakeStore, "coresend.dev", writeStaticFixture(t))

	t.Run("inbox route success", func(t *testing.T) {
		t.Parallel()

		req, address := newSignedRouteRequest(t, http.MethodGet, "/api/inbox/{address}", nil, time.Now())
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if fakeStore.getEmailsCallCount < 1 {
			t.Fatalf("expected getEmails to be called at least once")
		}
		if fakeStore.lastGetEmailsAddress != address {
			t.Fatalf("getEmails address = %q, want %q", fakeStore.lastGetEmailsAddress, address)
		}
	})

	t.Run("delete route success", func(t *testing.T) {
		t.Parallel()

		req, address := newSignedRouteRequest(t, http.MethodDelete, "/api/inbox/{address}/email-1", nil, time.Now())
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if fakeStore.deleteEmailCallCount < 1 {
			t.Fatalf("expected deleteEmail to be called at least once")
		}
		if fakeStore.lastDeleteAddressBox != address {
			t.Fatalf("deleteEmail address = %q, want %q", fakeStore.lastDeleteAddressBox, address)
		}
		if fakeStore.lastDeleteEmailID != "email-1" {
			t.Fatalf("deleteEmail id = %q, want %q", fakeStore.lastDeleteEmailID, "email-1")
		}
	})
}

func TestNewRouter_RateLimitEnforced_OnInboxAndDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		pathTmpl   string
		wantPrefix string
	}{
		{
			name:       "inbox route",
			method:     http.MethodGet,
			pathTmpl:   "/api/inbox/{address}",
			wantPrefix: "inbox",
		},
		{
			name:       "delete route",
			method:     http.MethodDelete,
			pathTmpl:   "/api/inbox/{address}/email-1",
			wantPrefix: "delete",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fakeStore := &fakeEmailStore{
				checkRateLimitFn: func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
					return false, 0, nil
				},
				getEmailsFn: func(ctx context.Context, addressBox string) ([]store.Email, error) {
					t.Fatalf("handler should not be called when rate limit blocks")
					return nil, nil
				},
				deleteEmailFn: func(ctx context.Context, addressBox string, emailID string) error {
					t.Fatalf("handler should not be called when rate limit blocks")
					return nil
				},
			}
			router := NewRouter(fakeStore, "coresend.dev", writeStaticFixture(t))

			req, _ := newSignedRouteRequest(t, tc.method, tc.pathTmpl, nil, time.Now())
			req.RemoteAddr = "192.0.2.10:7777"
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusTooManyRequests {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusTooManyRequests)
			}
			errResp := decodeErrorResponse(t, rr)
			if errResp.Error.Code != ErrCodeRateLimitExceeded {
				t.Fatalf("error.code = %q, want %q", errResp.Error.Code, ErrCodeRateLimitExceeded)
			}
			if !strings.HasPrefix(fakeStore.lastRateLimitKey, tc.wantPrefix+":") {
				t.Fatalf("rate limit key = %q, expected prefix %q", fakeStore.lastRateLimitKey, tc.wantPrefix+":")
			}
		})
	}
}

func TestNewRouter_HealthRoute_NoAuthRequired(t *testing.T) {
	t.Parallel()

	t.Run("redis connected", func(t *testing.T) {
		t.Parallel()

		router := NewRouter(&fakeEmailStore{}, "coresend.dev", writeStaticFixture(t))
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("redis disconnected", func(t *testing.T) {
		t.Parallel()

		fakeStore := &fakeEmailStore{
			pingFn: func(ctx context.Context) error {
				return fmt.Errorf("redis down")
			},
		}
		router := NewRouter(fakeStore, "coresend.dev", writeStaticFixture(t))
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
		}
	})
}

func TestNewRouter_DocsAndMetricsReachable(t *testing.T) {
	t.Parallel()

	router := NewRouter(&fakeEmailStore{}, "coresend.dev", writeStaticFixture(t))

	reqMetrics := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rrMetrics := httptest.NewRecorder()
	router.ServeHTTP(rrMetrics, reqMetrics)
	if rrMetrics.Code != http.StatusOK {
		t.Fatalf("/metrics status = %d, want %d", rrMetrics.Code, http.StatusOK)
	}

	reqDocs := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	rrDocs := httptest.NewRecorder()
	router.ServeHTTP(rrDocs, reqDocs)
	if rrDocs.Code != http.StatusOK &&
		rrDocs.Code != http.StatusMovedPermanently &&
		rrDocs.Code != http.StatusFound &&
		rrDocs.Code != http.StatusTemporaryRedirect &&
		rrDocs.Code != http.StatusPermanentRedirect {
		t.Fatalf("/docs/ status = %d, expected reachable (200 or redirect)", rrDocs.Code)
	}
}

func TestNewRouter_MethodMismatchAndUnknownPath(t *testing.T) {
	t.Parallel()

	router := NewRouter(&fakeEmailStore{}, "coresend.dev", writeStaticFixture(t))

	reqMismatch := httptest.NewRequest(http.MethodPost, "/api/health", nil)
	rrMismatch := httptest.NewRecorder()
	router.ServeHTTP(rrMismatch, reqMismatch)
	if rrMismatch.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/health status = %d, want %d", rrMismatch.Code, http.StatusMethodNotAllowed)
	}

	reqUnknown := httptest.NewRequest(http.MethodGet, "/api/not-registered", nil)
	rrUnknown := httptest.NewRecorder()
	router.ServeHTTP(rrUnknown, reqUnknown)
	if rrUnknown.Code != http.StatusOK {
		t.Fatalf("GET /api/not-registered status = %d, want %d", rrUnknown.Code, http.StatusOK)
	}
}
