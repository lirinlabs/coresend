package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fn-jakubkarp/coresend/internal/store"
)

type mockEmailStore struct {
	emails map[string][]store.Email
	rate   map[string]int
}

func newMockEmailStore() *mockEmailStore {
	return &mockEmailStore{
		emails: make(map[string][]store.Email),
		rate:   make(map[string]int),
	}
}

func (m *mockEmailStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
	m.emails[addressBox] = append(m.emails[addressBox], email)
	return nil
}

func (m *mockEmailStore) GetEmails(ctx context.Context, addressBox string) ([]store.Email, error) {
	return m.emails[addressBox], nil
}

func (m *mockEmailStore) GetEmail(ctx context.Context, addressBox string, emailID string) (*store.Email, error) {
	emails := m.emails[addressBox]
	for _, email := range emails {
		if email.ID == emailID {
			return &email, nil
		}
	}
	return nil, nil
}

func (m *mockEmailStore) DeleteEmail(ctx context.Context, addressBox string, emailID string) error {
	emails := m.emails[addressBox]
	for i, email := range emails {
		if email.ID == emailID {
			m.emails[addressBox] = append(emails[:i], emails[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockEmailStore) ClearInbox(ctx context.Context, addressBox string) (int64, error) {
	count := int64(len(m.emails[addressBox]))
	delete(m.emails, addressBox)
	return count, nil
}

func (m *mockEmailStore) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	m.rate[key]++
	remaining := limit - m.rate[key]
	return m.rate[key] <= limit, remaining, nil
}

func (m *mockEmailStore) Ping(ctx context.Context) error {
	return nil
}

func TestGenerateMnemonic(t *testing.T) {
	handler := NewAPIHandler(newMockEmailStore(), "example.com")

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantFields []string
	}{
		{
			name:       "valid request",
			method:     http.MethodPost,
			wantStatus: http.StatusOK,
			wantFields: []string{"mnemonic", "address", "email"},
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/identity/generate", nil)
			w := httptest.NewRecorder()

			handler.handleGenerateMnemonic(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var resp GenerateMnemonicResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				for _, field := range tt.wantFields {
					if field == "mnemonic" && resp.Mnemonic == "" {
						t.Errorf("Expected non-empty mnemonic")
					}
					if field == "address" && resp.Address == "" {
						t.Errorf("Expected non-empty address")
					}
					if field == "email" && resp.Email == "" {
						t.Errorf("Expected non-empty email")
					}
				}
			}
		})
	}
}

func TestDeriveAddress(t *testing.T) {
	handler := NewAPIHandler(newMockEmailStore(), "example.com")

	tests := []struct {
		name       string
		method     string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "valid request",
			method:     http.MethodPost,
			body:       DeriveAddressRequest{Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid request",
			method:     http.MethodPost,
			body:       "invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/api/identity/derive", &body)
			w := httptest.NewRecorder()

			handler.handleDeriveAddress(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestGetInbox(t *testing.T) {
	mockStore := newMockEmailStore()
	validAddr := "b4ebe3e2200cbc90"
	mockStore.emails[validAddr] = []store.Email{
		{ID: "1", From: "sender@example.com", Subject: "Test Email"},
	}
	handler := NewAPIHandler(mockStore, "example.com")

	tests := []struct {
		name       string
		path       string
		method     string
		wantStatus int
		wantCount  int
	}{
		{
			name:       "valid address",
			path:       "/api/inbox/" + validAddr,
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name:       "invalid address",
			path:       "/api/inbox/invalid",
			method:     http.MethodGet,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty address",
			path:       "/api/inbox/",
			method:     http.MethodGet,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.handleGetInbox(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var resp InboxResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if resp.Count != tt.wantCount {
					t.Errorf("Expected count %d, got %d", tt.wantCount, resp.Count)
				}
			}
		})
	}
}

func TestHealth(t *testing.T) {
	handler := NewAPIHandler(newMockEmailStore(), "example.com")

	tests := []struct {
		name       string
		method     string
		wantStatus int
	}{
		{
			name:       "valid request",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/health", nil)
			w := httptest.NewRecorder()

			handler.handleHealth(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
