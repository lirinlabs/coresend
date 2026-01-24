package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/fn-jakubkarp/coresend/internal/identity"
	"github.com/fn-jakubkarp/coresend/internal/store"
)

type APIHandler struct {
	Store  store.EmailStore
	Domain string
}

func NewAPIHandler(s store.EmailStore, domain string) *APIHandler {
	return &APIHandler{
		Store:  s,
		Domain: domain,
	}
}

func (h *APIHandler) handleGenerateMnemonic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mnemonic, err := identity.GenerateNewMnemonic()
	if err != nil {
		log.Printf("Error generating mnemonic: %v", err)
		writeError(w, ErrCodeInternalError, "Failed to generate mnemonic", http.StatusInternalServerError)
		return
	}

	address := identity.AddressFromMnemonic(mnemonic)
	email := address + "@" + h.Domain

	resp := GenerateMnemonicResponse{
		Mnemonic: mnemonic,
		Address:  address,
		Email:    email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleDeriveAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeriveAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrCodeInvalidMnemonic, "Invalid request body", http.StatusBadRequest)
		return
	}

	address := identity.AddressFromMnemonic(req.Mnemonic)
	email := address + "@" + h.Domain

	resp := DeriveAddressResponse{
		Address: address,
		Email:   email,
		Valid:   identity.IsValidBIP39Mnemonic(req.Mnemonic),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleValidateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/identity/validate/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, ErrCodeInvalidAddress, "Address is required", http.StatusBadRequest)
		return
	}

	address := parts[0]
	resp := ValidateAddressResponse{
		Address: address,
		Valid:   identity.IsValidAddress(address),
	}

	if !resp.Valid {
		resp.Reason = "Address must be exactly 16 hexadecimal characters"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleGetInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/inbox/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, ErrCodeInvalidAddress, "Address is required", http.StatusBadRequest)
		return
	}

	address := parts[0]
	if !identity.IsValidAddress(address) {
		writeErrorWithDetails(w, ErrCodeInvalidAddress, "Invalid address format", http.StatusBadRequest, map[string]interface{}{
			"provided":        address,
			"expected_length": 16,
		})
		return
	}

	emails, err := h.Store.GetEmails(r.Context(), address)
	if err != nil {
		log.Printf("Error getting emails: %v", err)
		writeError(w, ErrCodeInternalError, "Failed to retrieve emails", http.StatusInternalServerError)
		return
	}

	emailResponses := make([]EmailResponse, 0, len(emails))
	for _, email := range emails {
		emailResponses = append(emailResponses, EmailResponse{
			ID:         email.ID,
			From:       email.From,
			To:         email.To,
			Subject:    email.Subject,
			Body:       email.Body,
			ReceivedAt: email.ReceivedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	resp := InboxResponse{
		Address: address,
		Email:   address + "@" + h.Domain,
		Count:   len(emailResponses),
		Emails:  emailResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleGetEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/inbox/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, ErrCodeInvalidAddress, "Address and email ID are required", http.StatusBadRequest)
		return
	}

	address := parts[0]
	emailID := parts[1]

	if !identity.IsValidAddress(address) {
		writeErrorWithDetails(w, ErrCodeInvalidAddress, "Invalid address format", http.StatusBadRequest, map[string]interface{}{
			"provided":        address,
			"expected_length": 16,
		})
		return
	}

	email, err := h.Store.GetEmail(r.Context(), address, emailID)
	if err != nil {
		log.Printf("Error getting email: %v", err)
		writeError(w, ErrCodeInternalError, "Failed to retrieve email", http.StatusInternalServerError)
		return
	}

	if email == nil {
		writeError(w, ErrCodeNotFound, "Email not found", http.StatusNotFound)
		return
	}

	resp := EmailResponse{
		ID:         email.ID,
		From:       email.From,
		To:         email.To,
		Subject:    email.Subject,
		Body:       email.Body,
		ReceivedAt: email.ReceivedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleDeleteEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/inbox/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		writeError(w, ErrCodeInvalidAddress, "Address and email ID are required", http.StatusBadRequest)
		return
	}

	address := parts[0]
	emailID := parts[1]

	if !identity.IsValidAddress(address) {
		writeErrorWithDetails(w, ErrCodeInvalidAddress, "Invalid address format", http.StatusBadRequest, map[string]interface{}{
			"provided":        address,
			"expected_length": 16,
		})
		return
	}

	err := h.Store.DeleteEmail(r.Context(), address, emailID)
	if err != nil {
		log.Printf("Error deleting email: %v", err)
		writeError(w, ErrCodeInternalError, "Failed to delete email", http.StatusInternalServerError)
		return
	}

	resp := DeleteResponse{
		Deleted: true,
		ID:      emailID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleClearInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/inbox/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, ErrCodeInvalidAddress, "Address is required", http.StatusBadRequest)
		return
	}

	address := parts[0]
	if !identity.IsValidAddress(address) {
		writeErrorWithDetails(w, ErrCodeInvalidAddress, "Invalid address format", http.StatusBadRequest, map[string]interface{}{
			"provided":        address,
			"expected_length": 16,
		})
		return
	}

	count, err := h.Store.ClearInbox(r.Context(), address)
	if err != nil {
		log.Printf("Error clearing inbox: %v", err)
		writeError(w, ErrCodeInternalError, "Failed to clear inbox", http.StatusInternalServerError)
		return
	}

	resp := DeleteResponse{
		Deleted: true,
		Count:   count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, ErrCodeInternalError, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	redisStatus := "connected"
	if err := h.Store.Ping(r.Context()); err != nil {
		redisStatus = "disconnected"
	}

	resp := HealthResponse{
		Status: redisStatus,
		Services: map[string]string{
			"redis": redisStatus,
			"smtp":  "running",
		},
	}

	if redisStatus != "connected" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
