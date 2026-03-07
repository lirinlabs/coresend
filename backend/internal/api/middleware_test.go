package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/store"
)

type fakeEmailStore struct {
	saveEmailFn       func(ctx context.Context, addressBox string, email store.Email) error
	getEmailsFn       func(ctx context.Context, addressBox string) ([]store.Email, error)
	getEmailFn        func(ctx context.Context, addressBox string, emailID string) (*store.Email, error)
	deleteEmailFn     func(ctx context.Context, addressBox string, emailID string) error
	clearInboxFn      func(ctx context.Context, addressBox string) (int64, error)
	registerAddressFn func(ctx context.Context, addressBox string, duration time.Duration) error
	isAddressActiveFn func(ctx context.Context, addressBox string) (bool, error)
	pingFn            func(ctx context.Context) error

	checkRateLimitFn func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)
	checkNonceFn     func(ctx context.Context, nonce string, ttl time.Duration) (bool, error)

	lastSaveAddressBox string
	lastSavedEmail     store.Email
	saveEmailCallCount int

	lastGetEmailsAddress string
	getEmailsCallCount   int

	lastGetEmailAddress string
	lastGetEmailID      string
	getEmailCallCount   int

	lastDeleteAddressBox string
	lastDeleteEmailID    string
	deleteEmailCallCount int

	lastClearInboxAddress string
	clearInboxCallCount   int

	lastRegisterAddress  string
	lastRegisterDuration time.Duration
	registerCallCount    int

	lastIsAddressActiveAddress string
	isAddressActiveCallCount   int

	pingCallCount int

	lastRateLimitKey    string
	lastRateLimitLimit  int
	lastRateLimitWindow time.Duration

	lastNonce      string
	lastNonceTTL   time.Duration
	nonceCallCount int
}

func (f *fakeEmailStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
	f.lastSaveAddressBox = addressBox
	f.lastSavedEmail = email
	f.saveEmailCallCount++
	if f.saveEmailFn != nil {
		return f.saveEmailFn(ctx, addressBox, email)
	}
	return nil
}

func (f *fakeEmailStore) GetEmails(ctx context.Context, addressBox string) ([]store.Email, error) {
	f.lastGetEmailsAddress = addressBox
	f.getEmailsCallCount++
	if f.getEmailsFn != nil {
		return f.getEmailsFn(ctx, addressBox)
	}
	return nil, nil
}

func (f *fakeEmailStore) GetEmail(ctx context.Context, addressBox string, emailID string) (*store.Email, error) {
	f.lastGetEmailAddress = addressBox
	f.lastGetEmailID = emailID
	f.getEmailCallCount++
	if f.getEmailFn != nil {
		return f.getEmailFn(ctx, addressBox, emailID)
	}
	return nil, nil
}

func (f *fakeEmailStore) DeleteEmail(ctx context.Context, addressBox string, emailID string) error {
	f.lastDeleteAddressBox = addressBox
	f.lastDeleteEmailID = emailID
	f.deleteEmailCallCount++
	if f.deleteEmailFn != nil {
		return f.deleteEmailFn(ctx, addressBox, emailID)
	}
	return nil
}

func (f *fakeEmailStore) ClearInbox(ctx context.Context, addressBox string) (int64, error) {
	f.lastClearInboxAddress = addressBox
	f.clearInboxCallCount++
	if f.clearInboxFn != nil {
		return f.clearInboxFn(ctx, addressBox)
	}
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
	f.lastRegisterAddress = addressBox
	f.lastRegisterDuration = duration
	f.registerCallCount++
	if f.registerAddressFn != nil {
		return f.registerAddressFn(ctx, addressBox, duration)
	}
	return nil
}

func (f *fakeEmailStore) IsAddressActive(ctx context.Context, addressBox string) (bool, error) {
	f.lastIsAddressActiveAddress = addressBox
	f.isAddressActiveCallCount++
	if f.isAddressActiveFn != nil {
		return f.isAddressActiveFn(ctx, addressBox)
	}
	return true, nil
}

func (f *fakeEmailStore) Ping(ctx context.Context) error {
	f.pingCallCount++
	if f.pingFn != nil {
		return f.pingFn(ctx)
	}
	return nil
}

func (f *fakeEmailStore) CheckAndStoreNonce(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
	f.lastNonce = nonce
	f.lastNonceTTL = ttl
	f.nonceCallCount++
	if f.checkNonceFn == nil {
		return true, nil
	}
	return f.checkNonceFn(ctx, nonce, ttl)
}

func decodeErrorResponse(t *testing.T, rr *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()

	var got ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode error response body: %v", err)
	}
	return got
}

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
			name:  "8 char hex segment not replaced",
			input: "/api/inbox/deadbeef",
			want:  "/api/inbox/deadbeef",
		},
		{
			name:  "9 char hex segment replaced",
			input: "/api/inbox/deadbeef0",
			want:  "/api/inbox/{id}",
		},
		{
			name:  "9 char non hex segment not replaced",
			input: "/api/inbox/deadbeefg",
			want:  "/api/inbox/deadbeefg",
		},
		{
			name:  "32 char non hex segment replaced due length threshold",
			input: "/api/inbox/zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			want:  "/api/inbox/{id}",
		},
		{
			name:  "mixed static and dynamic segments",
			input: "/api/inbox/abc/0123456789abcdef0123456789abcdef",
			want:  "/api/inbox/abc/{id}",
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
		{
			name:  "empty path normalizes to root",
			input: "",
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

func TestIsHexOrAlphanumeric(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "lowercase hex", input: "deadbeef", want: true},
		{name: "uppercase hex", input: "DEADBEEF", want: true},
		{name: "numeric", input: "0123456789", want: true},
		{name: "mixed hex", input: "abcDEF123", want: true},
		{name: "contains dash", input: "abc-def", want: false},
		{name: "contains nonhex letter", input: "deadbeefg", want: false},
		{name: "contains whitespace", input: "dead beef", want: false},
		{name: "empty string", input: "", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isHexOrAlphanumeric(tc.input)
			if got != tc.want {
				t.Fatalf("isHexOrAlphanumeric(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Parallel()

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
		wantErrorCode    string
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
			wantStatus:     http.StatusTooManyRequests,
			wantNextCalled: false,
			wantErrorCode:  ErrCodeRateLimitExceeded,
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
			t.Parallel()

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

			if tc.wantErrorCode != "" {
				got := decodeErrorResponse(t, rr)
				if got.Error.Code != tc.wantErrorCode {
					t.Fatalf("error.code = %q, want %q", got.Error.Code, tc.wantErrorCode)
				}
			}
		})
	}
}
