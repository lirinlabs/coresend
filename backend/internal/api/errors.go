package api

import "encoding/json"
import "net/http"

const (
	ErrCodeInvalidAddress     = "INVALID_ADDRESS"
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

func writeErrorWithDetails(w http.ResponseWriter, code string, message string, httpStatus int, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	resp := ErrorResponse{
		Error: ErrorDetails{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	json.NewEncoder(w).Encode(resp)
}
