package api

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/identity"
	"golang.org/x/crypto/ed25519"
)

func TestAuthMiddleware(t *testing.T) {
	testMnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	privkey, pubkey, _ := identity.DeriveEd25519KeyPair(testMnemonic)
	testAddress := identity.AddressFromPublicKey(pubkey)

	timestamp := time.Now().UnixMilli()
	message := identity.CreateMessageToSign(testAddress, timestamp)
	signature := ed25519.Sign(privkey, []byte(message))
	timestampStr := strconv.FormatInt(timestamp, 10)

	tests := []struct {
		name           string
		authAddress    string
		authTimestamp  string
		authPubkey     string
		authSignature  string
		path           string
		expectedStatus int
	}{
		{
			name:           "valid authorization",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing X-Auth-Address",
			authAddress:    "",
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing X-Auth-Timestamp",
			authAddress:    testAddress,
			authTimestamp:  "",
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing X-Auth-Pubkey",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     "",
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing X-Auth-Signature",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  "",
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid address format",
			authAddress:    "invalid",
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid timestamp format",
			authAddress:    testAddress,
			authTimestamp:  "not-a-number",
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired timestamp",
			authAddress:    testAddress,
			authTimestamp:  strconv.FormatInt(time.Now().Add(-2*time.Minute).UnixMilli(), 10),
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "future timestamp",
			authAddress:    testAddress,
			authTimestamp:  strconv.FormatInt(time.Now().Add(2*time.Minute).UnixMilli(), 10),
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid public key format",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     "not-hex",
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid signature format",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  "not-hex",
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "address does not match public key",
			authAddress:    "0000000000000000",
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid signature",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString([]byte("invalid")),
			path:           "/api/inbox/" + testAddress,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "path address mismatch",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
			path:           "/api/inbox/0000000000000000",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid path",
			authAddress:    testAddress,
			authTimestamp:  timestampStr,
			authPubkey:     hex.EncodeToString(pubkey),
			authSignature:  hex.EncodeToString(signature),
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
			if tt.authTimestamp != "" {
				req.Header.Set("X-Auth-Timestamp", tt.authTimestamp)
			}
			if tt.authPubkey != "" {
				req.Header.Set("X-Auth-Pubkey", tt.authPubkey)
			}
			if tt.authSignature != "" {
				req.Header.Set("X-Auth-Signature", tt.authSignature)
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
