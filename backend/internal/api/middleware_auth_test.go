package api

import (
	"bytes"
	"context"
	"crypto/ed25519"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type signedRequestFixture struct {
	req       *http.Request
	address   string
	nonce     string
	timestamp string
	body      []byte
}

func newSignedRequestFixture(t *testing.T, method, path string, body []byte, at time.Time) signedRequestFixture {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 keypair: %v", err)
	}

	hash := sha256.Sum256(publicKey)
	address := hex.EncodeToString(hash[:])[:40]
	nonce := "550e8400-e29b-41d4-a716-446655440000"
	ts := strconv.FormatInt(at.Unix(), 10)
	bodyHash := sha256.Sum256(body)
	payload := fmt.Sprintf("%s:%s:%s:%s:%s", method, path, ts, hex.EncodeToString(bodyHash[:]), nonce)
	signature := ed25519.Sign(privateKey, []byte(payload))

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.SetPathValue("address", address)
	req.Header.Set("X-Public-Key", hex.EncodeToString(publicKey))
	req.Header.Set("X-Signature", hex.EncodeToString(signature))
	req.Header.Set("X-Timestamp", ts)
	req.Header.Set("X-Nonce", nonce)

	return signedRequestFixture{
		req:       req,
		address:   address,
		nonce:     nonce,
		timestamp: ts,
		body:      body,
	}
}

type errorReadCloser struct{}

func (errorReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("forced read error")
}

func (errorReadCloser) Close() error {
	return nil
}

func TestSignatureAuthMiddleware_Failures(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name       string
		mutate     func(f *signedRequestFixture, store *fakeEmailStore)
		wantStatus int
		wantCode   string
	}

	now := time.Now()

	tests := []testCase{
		{
			name: "missing headers",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header = http.Header{}
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "invalid nonce format",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Nonce", "not-a-uuid")
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "invalid timestamp format",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Timestamp", "not-a-number")
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "expired timestamp",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Timestamp", strconv.FormatInt(now.Add(-6*time.Minute).Unix(), 10))
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "future timestamp outside window",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Timestamp", strconv.FormatInt(now.Add(6*time.Minute).Unix(), 10))
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "invalid public key encoding",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Public-Key", "zz")
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "invalid public key size",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Public-Key", hex.EncodeToString(make([]byte, ed25519.PublicKeySize-1)))
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "invalid signature encoding",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Signature", "zz")
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "invalid signature size",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Header.Set("X-Signature", hex.EncodeToString(make([]byte, ed25519.SignatureSize-1)))
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "missing address path value",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.SetPathValue("address", "")
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "address mismatch",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.SetPathValue("address", strings.Repeat("a", 40))
			},
			wantStatus: http.StatusForbidden,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "request body read failure",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				f.req.Body = errorReadCloser{}
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   ErrCodeInternalError,
		},
		{
			name: "signature verification failure",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				sig, err := hex.DecodeString(f.req.Header.Get("X-Signature"))
				if err != nil {
					t.Fatalf("failed to decode signature: %v", err)
				}
				sig[0] ^= 0xFF
				f.req.Header.Set("X-Signature", hex.EncodeToString(sig))
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
		{
			name: "nonce store error",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				store.checkNonceFn = func(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
					return false, errors.New("redis unavailable")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   ErrCodeInternalError,
		},
		{
			name: "nonce reuse",
			mutate: func(f *signedRequestFixture, store *fakeEmailStore) {
				store.checkNonceFn = func(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
					return false, nil
				}
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   ErrCodeUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fixture := newSignedRequestFixture(t, http.MethodPost, "/api/inbox/resource", []byte(`{"ok":true}`), now)
			store := &fakeEmailStore{}
			tc.mutate(&fixture, store)

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			handler := signatureAuthMiddleware(store)(next)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, fixture.req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if nextCalled {
				t.Fatalf("next handler should not be called")
			}

			errResp := decodeErrorResponse(t, rr)
			if errResp.Error.Code != tc.wantCode {
				t.Fatalf("error.code = %q, want %q", errResp.Error.Code, tc.wantCode)
			}
		})
	}
}

func TestSignatureAuthMiddleware_Success(t *testing.T) {
	t.Parallel()

	body := []byte(`{"name":"CoreSend"}`)
	fixture := newSignedRequestFixture(t, http.MethodPost, "/api/inbox/resource", body, time.Now())

	store := &fakeEmailStore{
		checkNonceFn: func(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}

	var nextBody []byte
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true

		readBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body in next handler: %v", err)
		}
		nextBody = readBody
		w.WriteHeader(http.StatusNoContent)
	})

	handler := signatureAuthMiddleware(store)(next)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, fixture.req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Fatalf("expected next handler to be called")
	}
	if string(nextBody) != string(body) {
		t.Fatalf("body in next handler = %q, want %q", string(nextBody), string(body))
	}

	if store.lastNonce != fixture.nonce {
		t.Fatalf("nonce passed to store = %q, want %q", store.lastNonce, fixture.nonce)
	}
	if store.lastNonceTTL != 5*time.Minute {
		t.Fatalf("nonce ttl = %s, want %s", store.lastNonceTTL, 5*time.Minute)
	}
	if store.nonceCallCount != 1 {
		t.Fatalf("nonce store call count = %d, want %d", store.nonceCallCount, 1)
	}
}
