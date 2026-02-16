package api

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/fn-jakubkarp/coresend/internal/identity"
	"github.com/fn-jakubkarp/coresend/internal/store"

	_ "github.com/fn-jakubkarp/coresend/docs"
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

// @Summary Generate new identity
// @Description Generate a new BIP39 mnemonic phrase and derive an Ed25519 key pair with address
// @Tags identity
// @Accept json
// @Produce json
// @Success 200 {object} GenerateMnemonicResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/identity/generate [post]
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

	_, pubkey, err := identity.DeriveEd25519KeyPair(mnemonic)
	if err != nil {
		log.Printf("Error deriving keypair: %v", err)
		writeError(w, ErrCodeInternalError, "Failed to derive keys", http.StatusInternalServerError)
		return
	}

	address := identity.AddressFromPublicKey(pubkey)
	email := address + "@" + h.Domain

	resp := GenerateMnemonicResponse{
		Mnemonic:  mnemonic,
		Address:   address,
		PublicKey: hex.EncodeToString(pubkey),
		Email:     email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// @Summary Derive address from mnemonic
// @Description Derive an address and public key from an existing BIP39 mnemonic phrase
// @Tags identity
// @Accept json
// @Produce json
// @Param request body DeriveAddressRequest true "Mnemonic phrase"
// @Success 200 {object} DeriveAddressResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/identity/derive [post]
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

	valid := identity.IsValidBIP39Mnemonic(req.Mnemonic)
	var pubkey []byte
	var address string

	if valid {
		_, pubkey, _ = identity.DeriveEd25519KeyPair(req.Mnemonic)
		address = identity.AddressFromPublicKey(pubkey)
	}

	email := address + "@" + h.Domain

	resp := DeriveAddressResponse{
		Address:   address,
		Email:     email,
		PublicKey: hex.EncodeToString(pubkey),
		Valid:     valid,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// @Summary Validate address format
// @Description Check if an address is valid (16 hexadecimal characters)
// @Tags identity
// @Accept json
// @Produce json
// @Param address path string true "Address to validate"
// @Success 200 {object} ValidateAddressResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/identity/validate/{address} [get]
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

// @Summary Get inbox emails
// @Description Retrieve all emails for a specific address
// @Tags inbox
// @Accept json
// @Produce json
// @Param address path string true "Address to retrieve emails for"
// @Success 200 {object} InboxResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security SignatureAuth
// @Router /api/inbox/{address} [get]
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

// @Summary Get single email
// @Description Retrieve a specific email by ID for an address
// @Tags inbox
// @Accept json
// @Produce json
// @Param address path string true "Address"
// @Param emailId path string true "Email ID"
// @Success 200 {object} EmailResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security SignatureAuth
// @Router /api/inbox/{address}/{emailId} [get]
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

// @Summary Delete single email
// @Description Delete a specific email by ID for an address
// @Tags inbox
// @Accept json
// @Produce json
// @Param address path string true "Address"
// @Param emailId path string true "Email ID"
// @Success 200 {object} DeleteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security SignatureAuth
// @Router /api/inbox/{address}/{emailId} [delete]
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

// @Summary Clear entire inbox
// @Description Delete all emails for a specific address
// @Tags inbox
// @Accept json
// @Produce json
// @Success 200 {object} DeleteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security SignatureAuth
// @Router /api/inbox [delete]
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

// @Summary Health check
// @Description Check API and services health status
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /api/health [get]
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
