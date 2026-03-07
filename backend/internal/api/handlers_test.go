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

const testValidAddress = "0123456789abcdef0123456789abcdef01234567"

func decodeJSONResponse[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var got T
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode json response: %v", err)
	}
	return got
}

func TestNewAPIHandler(t *testing.T) {
	t.Parallel()

	s := &fakeEmailStore{}
	h := NewAPIHandler(s, "coresend.io")
	if h == nil {
		t.Fatalf("NewAPIHandler returned nil")
	}
	if h.Store != s {
		t.Fatalf("handler store was not assigned")
	}
	if h.Domain != "coresend.io" {
		t.Fatalf("handler domain = %q, want %q", h.Domain, "coresend.io")
	}
}

func TestHandleRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		address           string
		registerErr       error
		wantStatus        int
		wantErrorCode     string
		wantRegisterCalls int
	}{
		{
			name:              "missing address",
			address:           "",
			wantStatus:        http.StatusBadRequest,
			wantErrorCode:     ErrCodeInvalidAddress,
			wantRegisterCalls: 0,
		},
		{
			name:              "invalid address",
			address:           "abc",
			wantStatus:        http.StatusBadRequest,
			wantErrorCode:     ErrCodeInvalidAddress,
			wantRegisterCalls: 0,
		},
		{
			name:              "store error",
			address:           testValidAddress,
			registerErr:       errors.New("redis down"),
			wantStatus:        http.StatusInternalServerError,
			wantErrorCode:     ErrCodeInternalError,
			wantRegisterCalls: 1,
		},
		{
			name:              "success",
			address:           testValidAddress,
			wantStatus:        http.StatusOK,
			wantRegisterCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &fakeEmailStore{
				registerAddressFn: func(ctx context.Context, addressBox string, duration time.Duration) error {
					return tc.registerErr
				},
			}

			h := NewAPIHandler(s, "coresend.io")
			req := httptest.NewRequest(http.MethodPost, "/api/register/"+tc.address, nil)
			if tc.address != "" {
				req.SetPathValue("address", tc.address)
			}
			rr := httptest.NewRecorder()

			h.handleRegister(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if s.registerCallCount != tc.wantRegisterCalls {
				t.Fatalf("register call count = %d, want %d", s.registerCallCount, tc.wantRegisterCalls)
			}

			if tc.wantRegisterCalls > 0 {
				if s.lastRegisterAddress != tc.address {
					t.Fatalf("register address = %q, want %q", s.lastRegisterAddress, tc.address)
				}
				if s.lastRegisterDuration != 24*time.Hour {
					t.Fatalf("register ttl = %s, want %s", s.lastRegisterDuration, 24*time.Hour)
				}
			}

			if tc.wantErrorCode != "" {
				gotErr := decodeErrorResponse(t, rr)
				if gotErr.Error.Code != tc.wantErrorCode {
					t.Fatalf("error.code = %q, want %q", gotErr.Error.Code, tc.wantErrorCode)
				}
				return
			}

			if got := rr.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want %q", got, "application/json")
			}
			gotResp := decodeJSONResponse[RegisterResponse](t, rr)
			if !gotResp.Registered {
				t.Fatalf("registered = %v, want true", gotResp.Registered)
			}
			if gotResp.Address != tc.address {
				t.Fatalf("address = %q, want %q", gotResp.Address, tc.address)
			}
			if gotResp.ExpiresIn != int((24 * time.Hour).Seconds()) {
				t.Fatalf("expires_in = %d, want %d", gotResp.ExpiresIn, int((24 * time.Hour).Seconds()))
			}
		})
	}
}

func TestHandleGetInbox(t *testing.T) {
	t.Parallel()

	mailTimeA := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	mailTimeB := time.Date(2024, time.January, 3, 4, 5, 6, 0, time.UTC)

	tests := []struct {
		name              string
		address           string
		storeEmails       []store.Email
		storeErr          error
		wantStatus        int
		wantErrorCode     string
		wantGetEmailsCall int
	}{
		{
			name:              "missing address",
			address:           "",
			wantStatus:        http.StatusBadRequest,
			wantErrorCode:     ErrCodeInvalidAddress,
			wantGetEmailsCall: 0,
		},
		{
			name:              "store error",
			address:           testValidAddress,
			storeErr:          errors.New("redis down"),
			wantStatus:        http.StatusInternalServerError,
			wantErrorCode:     ErrCodeInternalError,
			wantGetEmailsCall: 1,
		},
		{
			name:              "success empty inbox",
			address:           testValidAddress,
			storeEmails:       []store.Email{},
			wantStatus:        http.StatusOK,
			wantGetEmailsCall: 1,
		},
		{
			name:    "success with emails",
			address: testValidAddress,
			storeEmails: []store.Email{
				{
					ID:         "id-1",
					From:       "alice@example.com",
					To:         []string{testValidAddress + "@coresend.io"},
					Subject:    "Subject 1",
					Body:       "Body 1",
					ReceivedAt: mailTimeA,
				},
				{
					ID:         "id-2",
					From:       "bob@example.com",
					To:         []string{testValidAddress + "@coresend.io"},
					Subject:    "Subject 2",
					Body:       "Body 2",
					ReceivedAt: mailTimeB,
				},
			},
			wantStatus:        http.StatusOK,
			wantGetEmailsCall: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &fakeEmailStore{
				getEmailsFn: func(ctx context.Context, addressBox string) ([]store.Email, error) {
					return tc.storeEmails, tc.storeErr
				},
			}
			h := NewAPIHandler(s, "coresend.io")

			req := httptest.NewRequest(http.MethodGet, "/api/inbox/"+tc.address, nil)
			if tc.address != "" {
				req.SetPathValue("address", tc.address)
			}
			rr := httptest.NewRecorder()

			h.handleGetInbox(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if s.getEmailsCallCount != tc.wantGetEmailsCall {
				t.Fatalf("getEmails call count = %d, want %d", s.getEmailsCallCount, tc.wantGetEmailsCall)
			}
			if tc.wantGetEmailsCall > 0 && s.lastGetEmailsAddress != tc.address {
				t.Fatalf("getEmails address = %q, want %q", s.lastGetEmailsAddress, tc.address)
			}

			if tc.wantErrorCode != "" {
				gotErr := decodeErrorResponse(t, rr)
				if gotErr.Error.Code != tc.wantErrorCode {
					t.Fatalf("error.code = %q, want %q", gotErr.Error.Code, tc.wantErrorCode)
				}
				return
			}

			if got := rr.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want %q", got, "application/json")
			}

			resp := decodeJSONResponse[InboxResponse](t, rr)
			if resp.Address != tc.address {
				t.Fatalf("address = %q, want %q", resp.Address, tc.address)
			}
			if resp.Email != tc.address+"@coresend.io" {
				t.Fatalf("email = %q, want %q", resp.Email, tc.address+"@coresend.io")
			}
			if resp.Count != len(tc.storeEmails) {
				t.Fatalf("count = %d, want %d", resp.Count, len(tc.storeEmails))
			}
			if len(resp.Emails) != len(tc.storeEmails) {
				t.Fatalf("emails length = %d, want %d", len(resp.Emails), len(tc.storeEmails))
			}

			for i := range tc.storeEmails {
				if resp.Emails[i].ID != tc.storeEmails[i].ID {
					t.Fatalf("emails[%d].id = %q, want %q", i, resp.Emails[i].ID, tc.storeEmails[i].ID)
				}
				wantTs := tc.storeEmails[i].ReceivedAt.Format("2006-01-02T15:04:05Z")
				if resp.Emails[i].ReceivedAt != wantTs {
					t.Fatalf("emails[%d].received_at = %q, want %q", i, resp.Emails[i].ReceivedAt, wantTs)
				}
			}
		})
	}
}

func TestHandleGetEmail(t *testing.T) {
	t.Parallel()

	mailTime := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name            string
		address         string
		emailID         string
		storeEmail      *store.Email
		storeErr        error
		wantStatus      int
		wantErrorCode   string
		wantGetEmailRun int
	}{
		{
			name:            "missing params",
			address:         "",
			emailID:         "",
			wantStatus:      http.StatusBadRequest,
			wantErrorCode:   ErrCodeInvalidAddress,
			wantGetEmailRun: 0,
		},
		{
			name:            "store error",
			address:         testValidAddress,
			emailID:         "email-1",
			storeErr:        errors.New("redis down"),
			wantStatus:      http.StatusInternalServerError,
			wantErrorCode:   ErrCodeInternalError,
			wantGetEmailRun: 1,
		},
		{
			name:            "not found",
			address:         testValidAddress,
			emailID:         "email-1",
			storeEmail:      nil,
			wantStatus:      http.StatusNotFound,
			wantErrorCode:   ErrCodeNotFound,
			wantGetEmailRun: 1,
		},
		{
			name:    "success",
			address: testValidAddress,
			emailID: "email-1",
			storeEmail: &store.Email{
				ID:         "email-1",
				From:       "alice@example.com",
				To:         []string{testValidAddress + "@coresend.io"},
				Subject:    "Subject",
				Body:       "Body",
				ReceivedAt: mailTime,
			},
			wantStatus:      http.StatusOK,
			wantGetEmailRun: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &fakeEmailStore{
				getEmailFn: func(ctx context.Context, addressBox string, emailID string) (*store.Email, error) {
					return tc.storeEmail, tc.storeErr
				},
			}
			h := NewAPIHandler(s, "coresend.io")

			req := httptest.NewRequest(http.MethodGet, "/api/inbox/"+tc.address+"/"+tc.emailID, nil)
			if tc.address != "" {
				req.SetPathValue("address", tc.address)
			}
			if tc.emailID != "" {
				req.SetPathValue("emailId", tc.emailID)
			}
			rr := httptest.NewRecorder()

			h.handleGetEmail(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if s.getEmailCallCount != tc.wantGetEmailRun {
				t.Fatalf("getEmail call count = %d, want %d", s.getEmailCallCount, tc.wantGetEmailRun)
			}
			if tc.wantGetEmailRun > 0 {
				if s.lastGetEmailAddress != tc.address {
					t.Fatalf("getEmail address = %q, want %q", s.lastGetEmailAddress, tc.address)
				}
				if s.lastGetEmailID != tc.emailID {
					t.Fatalf("getEmail id = %q, want %q", s.lastGetEmailID, tc.emailID)
				}
			}

			if tc.wantErrorCode != "" {
				gotErr := decodeErrorResponse(t, rr)
				if gotErr.Error.Code != tc.wantErrorCode {
					t.Fatalf("error.code = %q, want %q", gotErr.Error.Code, tc.wantErrorCode)
				}
				return
			}

			resp := decodeJSONResponse[EmailResponse](t, rr)
			if resp.ID != tc.storeEmail.ID {
				t.Fatalf("id = %q, want %q", resp.ID, tc.storeEmail.ID)
			}
			wantTs := tc.storeEmail.ReceivedAt.Format("2006-01-02T15:04:05Z")
			if resp.ReceivedAt != wantTs {
				t.Fatalf("received_at = %q, want %q", resp.ReceivedAt, wantTs)
			}
		})
	}
}

func TestHandleDeleteEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		address          string
		emailID          string
		deleteErr        error
		wantStatus       int
		wantErrorCode    string
		wantDeleteCalled int
	}{
		{
			name:             "missing params",
			address:          "",
			emailID:          "",
			wantStatus:       http.StatusBadRequest,
			wantErrorCode:    ErrCodeInvalidAddress,
			wantDeleteCalled: 0,
		},
		{
			name:             "store error",
			address:          testValidAddress,
			emailID:          "email-1",
			deleteErr:        errors.New("redis down"),
			wantStatus:       http.StatusInternalServerError,
			wantErrorCode:    ErrCodeInternalError,
			wantDeleteCalled: 1,
		},
		{
			name:             "success",
			address:          testValidAddress,
			emailID:          "email-1",
			wantStatus:       http.StatusOK,
			wantDeleteCalled: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &fakeEmailStore{
				deleteEmailFn: func(ctx context.Context, addressBox string, emailID string) error {
					return tc.deleteErr
				},
			}
			h := NewAPIHandler(s, "coresend.io")

			req := httptest.NewRequest(http.MethodDelete, "/api/inbox/"+tc.address+"/"+tc.emailID, nil)
			if tc.address != "" {
				req.SetPathValue("address", tc.address)
			}
			if tc.emailID != "" {
				req.SetPathValue("emailId", tc.emailID)
			}
			rr := httptest.NewRecorder()

			h.handleDeleteEmail(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if s.deleteEmailCallCount != tc.wantDeleteCalled {
				t.Fatalf("deleteEmail call count = %d, want %d", s.deleteEmailCallCount, tc.wantDeleteCalled)
			}
			if tc.wantDeleteCalled > 0 {
				if s.lastDeleteAddressBox != tc.address {
					t.Fatalf("deleteEmail address = %q, want %q", s.lastDeleteAddressBox, tc.address)
				}
				if s.lastDeleteEmailID != tc.emailID {
					t.Fatalf("deleteEmail id = %q, want %q", s.lastDeleteEmailID, tc.emailID)
				}
			}

			if tc.wantErrorCode != "" {
				gotErr := decodeErrorResponse(t, rr)
				if gotErr.Error.Code != tc.wantErrorCode {
					t.Fatalf("error.code = %q, want %q", gotErr.Error.Code, tc.wantErrorCode)
				}
				return
			}

			resp := decodeJSONResponse[DeleteResponse](t, rr)
			if !resp.Deleted {
				t.Fatalf("deleted = %v, want true", resp.Deleted)
			}
			if resp.ID != tc.emailID {
				t.Fatalf("id = %q, want %q", resp.ID, tc.emailID)
			}
		})
	}
}

func TestHandleClearInbox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		address         string
		clearCount      int64
		clearErr        error
		wantStatus      int
		wantErrorCode   string
		wantClearCalled int
	}{
		{
			name:            "missing address",
			address:         "",
			wantStatus:      http.StatusBadRequest,
			wantErrorCode:   ErrCodeInvalidAddress,
			wantClearCalled: 0,
		},
		{
			name:            "store error",
			address:         testValidAddress,
			clearErr:        errors.New("redis down"),
			wantStatus:      http.StatusInternalServerError,
			wantErrorCode:   ErrCodeInternalError,
			wantClearCalled: 1,
		},
		{
			name:            "success",
			address:         testValidAddress,
			clearCount:      2,
			wantStatus:      http.StatusOK,
			wantClearCalled: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &fakeEmailStore{
				clearInboxFn: func(ctx context.Context, addressBox string) (int64, error) {
					return tc.clearCount, tc.clearErr
				},
			}
			h := NewAPIHandler(s, "coresend.io")

			req := httptest.NewRequest(http.MethodDelete, "/api/inbox/"+tc.address, nil)
			if tc.address != "" {
				req.SetPathValue("address", tc.address)
			}
			rr := httptest.NewRecorder()

			h.handleClearInbox(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if s.clearInboxCallCount != tc.wantClearCalled {
				t.Fatalf("clearInbox call count = %d, want %d", s.clearInboxCallCount, tc.wantClearCalled)
			}
			if tc.wantClearCalled > 0 && s.lastClearInboxAddress != tc.address {
				t.Fatalf("clearInbox address = %q, want %q", s.lastClearInboxAddress, tc.address)
			}

			if tc.wantErrorCode != "" {
				gotErr := decodeErrorResponse(t, rr)
				if gotErr.Error.Code != tc.wantErrorCode {
					t.Fatalf("error.code = %q, want %q", gotErr.Error.Code, tc.wantErrorCode)
				}
				return
			}

			resp := decodeJSONResponse[DeleteResponse](t, rr)
			if !resp.Deleted {
				t.Fatalf("deleted = %v, want true", resp.Deleted)
			}
			if resp.Count != tc.clearCount {
				t.Fatalf("count = %d, want %d", resp.Count, tc.clearCount)
			}
		})
	}
}

func TestHandleHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantState  string
	}{
		{
			name:       "redis connected",
			wantStatus: http.StatusOK,
			wantState:  "connected",
		},
		{
			name:       "redis disconnected",
			pingErr:    errors.New("redis down"),
			wantStatus: http.StatusServiceUnavailable,
			wantState:  "disconnected",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &fakeEmailStore{
				pingFn: func(ctx context.Context) error {
					return tc.pingErr
				},
			}
			h := NewAPIHandler(s, "coresend.io")

			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			rr := httptest.NewRecorder()

			h.handleHealth(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if s.pingCallCount != 1 {
				t.Fatalf("ping call count = %d, want 1", s.pingCallCount)
			}
			if got := rr.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want %q", got, "application/json")
			}

			resp := decodeJSONResponse[HealthResponse](t, rr)
			if resp.Status != tc.wantState {
				t.Fatalf("status = %q, want %q", resp.Status, tc.wantState)
			}
			if resp.Services["redis"] != tc.wantState {
				t.Fatalf("services.redis = %q, want %q", resp.Services["redis"], tc.wantState)
			}
			if resp.Services["smtp"] != "running" {
				t.Fatalf("services.smtp = %q, want %q", resp.Services["smtp"], "running")
			}
		})
	}
}
