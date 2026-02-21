package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/fn-jakubkarp/coresend/docs"
	"github.com/fn-jakubkarp/coresend/internal/store"
	"github.com/fn-jakubkarp/coresend/internal/validator"
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

// @Summary Register address for inbound mail
// @Description Register a derived address to actively receive emails for the next 24 hours
// @Tags inbox
// @Accept json
// @Produce json
// @Param address path string true "Address to register"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security SignatureAuth
// @Router /api/register/{address} [post]
func (h *APIHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	address := r.PathValue("address")
	if address == "" {
		writeError(w, ErrCodeInvalidAddress, "Address is required", http.StatusBadRequest)
		return
	}

	if !validator.IsValidHexAddress(address) {
		writeError(w, ErrCodeInvalidAddress, "Invalid address format", http.StatusBadRequest)
		return
	}

	ttl := 24 * time.Hour

	err := h.Store.RegisterAddress(r.Context(), address, ttl)
	if err != nil {
		log.Printf("Error registering address: %v", err)
		writeError(w, ErrCodeInternalError, "Failed to register address", http.StatusInternalServerError)
		return
	}

	resp := RegisterResponse{
		Registered: true,
		Address:    address,
		ExpiresIn:  int(ttl.Seconds()),
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
	address := r.PathValue("address")
	if address == "" {
		writeError(w, ErrCodeInvalidAddress, "Address is required", http.StatusBadRequest)
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

	address := r.PathValue("address")
	emailID := r.PathValue("emailId")

	if address == "" || emailID == "" {
		writeError(w, ErrCodeInvalidAddress, "Address and email ID are required", http.StatusBadRequest)
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
	address := r.PathValue("address")
	emailID := r.PathValue("emailId")

	if address == "" || emailID == "" {
		writeError(w, ErrCodeInvalidAddress, "Address and email ID are required", http.StatusBadRequest)
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
	address := r.PathValue("address")

	if address == "" {
		writeError(w, ErrCodeInvalidAddress, "Address is required", http.StatusBadRequest)
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
