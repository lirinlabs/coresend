package api

import (
	"encoding/json"
	"net/http"
)

const (
	ErrCodeInvalidAddress     = "INVALID_ADDRESS"
	ErrCodeInvalidMnemonic    = "INVALID_MNEMONIC"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
)

func writeError(w http.ResponseWriter, code string, message string, httpStatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	resp := ErrorResponse{
		Error: ErrorDetails{
			Code:    code,
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(resp)
}
