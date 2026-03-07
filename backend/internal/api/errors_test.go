package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()

	wantCode := ErrCodeInvalidAddress
	wantMessage := "Invalid address format"
	wantStatus := http.StatusBadRequest

	writeError(rr, wantCode, wantMessage, wantStatus)

	res := rr.Result()

	if res.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", res.StatusCode, wantStatus)
	}

	if got := res.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	var gotBody ErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&gotBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if gotBody.Error.Code != wantCode {
		t.Fatalf("error.code = %q, want %q", gotBody.Error.Code, wantCode)
	}

	if gotBody.Error.Message != wantMessage {
		t.Fatalf("error.message = %q, want %q", gotBody.Error.Message, wantMessage)
	}
}
