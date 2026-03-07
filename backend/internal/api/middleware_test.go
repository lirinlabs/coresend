package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/store"
)

func TestNormalizeEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "static health path unchanged",
			input: "/api/health",
			want:  "/api/health",
		},
		{
			name:  "40 char hex segment replaced",
			input: "/api/inbox/0123456789abcdef0123456789abcdef01234567",
			want:  "/api/inbox/{id}",
		},
		{
			name:  "uuid like segment replaced",
			input: "/api/inbox/550e8400-e29b-41d4-a716-446655440000",
			want:  "/api/inbox/{id}",
		},
		{
			name:  "non id segment unchanged",
			input: "/api/inbox/abc-def",
			want:  "/api/inbox/abc-def",
		},
		{
			name:  "root path",
			input: "/",
			want:  "/",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeEndpoint(tc.input)
			if got != tc.want {
				t.Fatalf("normalizeEndpoint(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

type fakeEmailStore struct {
	checkRateLimitFn func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)

	lastRateLimitKey    string
	lastRateLimitLimit  int
	lastRateLimitWindow time.Duration
}

func (f *fakeEmailStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
	return nil
}

func (f *fakeEmailStore) GetEmails(ctx context.Context, addressBox string) ([]store.Email, error) {
	return nil, nil
}

func (f *fakeEmailStore) GetEmail(ctx context.Context, addressBox string, emailID string) (*store.Email, error) {
	return nil, nil
}

func (f *fakeEmailStore) DeleteEmail(ctx context.Context, addressBox string, emailID string) error {
	return nil
}

func (f *fakeEmailStore) ClearInbox(ctx context.Context, addressBox string) (int64, error) {
	return 0, nil
}

func (f *fakeEmailStore) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	f.lastRateLimitKey = key
	f.lastRateLimitLimit = limit
	f.lastRateLimitWindow = window

	if f.checkRateLimitFn == nil {
		return true, 0, nil
	}
	return f.checkRateLimitFn(ctx, key, limit, window)
}

func (f *fakeEmailStore) RegisterAddress(ctx context.Context, addressBox string, duration time.Duration) error {
	return nil
}

func (f *fakeEmailStore) IsAddressActive(ctx context.Context, addressBox string) (bool, error) {
	return true, nil
}

func (f *fakeEmailStore) Ping(ctx context.Context) error {
	return nil
}

func (f *fakeEmailStore) CheckAndStoreNonce(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
	return true, nil
}

func TestRateLimitMiddleware(t *testing.T) {
	config := RateLimitConfig{
		Limit:     60,
		Window:    time.Minute,
		KeyPrefix: "inbox",
	}
	remoteAddr := "192.0.2.10:12345"
	expectedRateLimitKey := "inbox:" + remoteAddr

	tests := []struct {
		name             string
		checkRateLimitFn func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)
		wantStatus       int
		wantNextCalled   bool
		wantBodyContains string
	}{
		{
			name: "allowed calls next",
			checkRateLimitFn: func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
				return true, 10, nil
			},
			wantStatus:     http.StatusNoContent,
			wantNextCalled: true,
		},
		{
			name: "blocked returns 429",
			checkRateLimitFn: func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
				return false, 0, nil
			},
			wantStatus:       http.StatusTooManyRequests,
			wantNextCalled:   false,
			wantBodyContains: ErrCodeRateLimitExceeded,
		},
		{
			name: "store error calls next",
			checkRateLimitFn: func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
				return false, 0, errors.New("redis down")
			},
			wantStatus:     http.StatusNoContent,
			wantNextCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeStore := &fakeEmailStore{
				checkRateLimitFn: tc.checkRateLimitFn,
			}

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			handler := rateLimitMiddleware(fakeStore, config)(next)

			req := httptest.NewRequest(http.MethodGet, "/api/inbox/address", nil)
			req.RemoteAddr = remoteAddr
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}

			if nextCalled != tc.wantNextCalled {
				t.Fatalf("nextCalled = %v, want %v", nextCalled, tc.wantNextCalled)
			}

			if fakeStore.lastRateLimitKey != expectedRateLimitKey {
				t.Fatalf("rate limit key = %q, want %q", fakeStore.lastRateLimitKey, expectedRateLimitKey)
			}

			if fakeStore.lastRateLimitLimit != config.Limit {
				t.Fatalf("rate limit limit = %d, want %d", fakeStore.lastRateLimitLimit, config.Limit)
			}

			if fakeStore.lastRateLimitWindow != config.Window {
				t.Fatalf("rate limit window = %s, want %s", fakeStore.lastRateLimitWindow, config.Window)
			}

			if tc.wantBodyContains != "" && !strings.Contains(rr.Body.String(), tc.wantBodyContains) {
				t.Fatalf("response body = %q, expected to contain %q", rr.Body.String(), tc.wantBodyContains)
			}
		})
	}
}
