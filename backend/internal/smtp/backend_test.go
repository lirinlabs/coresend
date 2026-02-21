package smtp

import (
	"context"
	"strings"
	"testing"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/fn-jakubkarp/coresend/internal/store"
)

type mockStore struct {
	savedEmails []store.Email
	savedTo     []string
}

func (m *mockStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
	m.savedEmails = append(m.savedEmails, email)
	m.savedTo = append(m.savedTo, addressBox)
	return nil
}

func (m *mockStore) GetEmails(ctx context.Context, addressBox string) ([]store.Email, error) {
	return m.savedEmails, nil
}

func (m *mockStore) GetEmail(ctx context.Context, addressBox string, emailID string) (*store.Email, error) {
	return nil, nil
}

func (m *mockStore) DeleteEmail(ctx context.Context, addressBox string, emailID string) error {
	return nil
}

func (m *mockStore) ClearInbox(ctx context.Context, addressBox string) (int64, error) {
	return 0, nil
}

func (m *mockStore) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	return true, limit - 1, nil
}

func (m *mockStore) Ping(ctx context.Context) error {
	return nil
}

func TestBackend_NewSession(t *testing.T) {
	mockStore := &mockStore{}
	backend := &Backend{Store: mockStore}

	conn := &gosmtp.Conn{}
	session, err := backend.NewSession(conn)

	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	if session == nil {
		t.Fatal("NewSession() returned nil session")
	}

	smtpSession, ok := session.(*Session)
	if !ok {
		t.Fatal("NewSession() did not return *Session type")
	}

	if smtpSession.Store == nil {
		t.Error("NewSession() session store is nil")
	}
}

func TestSession_Mail(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{Store: mockStore}

	from := "sender@example.com"
	err := session.Mail(from, nil)

	if err != nil {
		t.Errorf("Mail() error = %v", err)
	}

	if session.From != from {
		t.Errorf("Mail() From = %v, want %v", session.From, from)
	}
}

func TestSession_Rcpt(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{Store: mockStore}

	to := "b4ebe3e2200cbc901234567890abcdef01234567@example.com"
	err := session.Rcpt(to, nil)

	if err != nil {
		t.Errorf("Rcpt() error = %v", err)
	}

	if len(session.To) != 1 || session.To[0] != "b4ebe3e2200cbc901234567890abcdef01234567" {
		t.Errorf("Rcpt() To = %v, want [b4ebe3e2200cbc901234567890abcdef01234567]", session.To)
	}
}

func TestSession_Rcpt_Multiple(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{Store: mockStore}

	recipients := []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@example.com",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb@example.com",
		"cccccccccccccccccccccccccccccccccccccccc@example.com",
	}
	for _, to := range recipients {
		err := session.Rcpt(to, nil)
		if err != nil {
			t.Errorf("Rcpt() error = %v", err)
		}
	}

	if len(session.To) != 3 {
		t.Errorf("Rcpt() To has %d recipients, want 3", len(session.To))
	}

	expectedLocalParts := []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "cccccccccccccccccccccccccccccccccccccccc"}
	for i, expected := range expectedLocalParts {
		if session.To[i] != expected {
			t.Errorf("Rcpt() To[%d] = %v, want %v", i, session.To[i], expected)
		}
	}
}

func TestSession_Rcpt_InvalidAddress(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{Store: mockStore}

	invalidAddresses := []string{
		"invalid@example.com",
		"admin@example.com",
		"test@example.com",
		"tooshort@example.com",
	}

	for _, to := range invalidAddresses {
		err := session.Rcpt(to, nil)
		if err == nil {
			t.Errorf("Rcpt(%s) expected error for invalid address, got nil", to)
		}
	}

	if len(session.To) != 0 {
		t.Errorf("Rcpt() should not have added invalid addresses, got %v", session.To)
	}
}

func TestSession_Reset(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{
		Store: mockStore,
		From:  "sender@example.com",
		To:    []string{"b4ebe3e2200cbc901234567890abcdef01234567"},
	}

	session.Reset()

	if session.From != "" {
		t.Errorf("Reset() From = %v, want empty", session.From)
	}

	if session.To != nil {
		t.Errorf("Reset() To = %v, want nil", session.To)
	}
}

func TestSession_Logout(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{Store: mockStore}

	err := session.Logout()

	if err != nil {
		t.Errorf("Logout() error = %v", err)
	}
}

func TestSession_Data(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{
		Store: mockStore,
		From:  "sender@example.com",
		To:    []string{"b4ebe3e2200cbc901234567890abcdef01234567"},
	}

	emailData := `From: sender@example.com
To: b4ebe3e2200cbc901234567890abcdef01234567@example.com
Subject: Test Subject
Content-Type: text/plain; charset=UTF-8

Test email body`

	err := session.Data(strings.NewReader(emailData))
	if err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	if len(mockStore.savedEmails) != 1 {
		t.Fatalf("Data() saved %d emails, want 1", len(mockStore.savedEmails))
	}

	savedEmail := mockStore.savedEmails[0]
	if savedEmail.From != "sender@example.com" {
		t.Errorf("Data() saved From = %v, want %v", savedEmail.From, "sender@example.com")
	}

	if len(savedEmail.To) != 1 || savedEmail.To[0] != "b4ebe3e2200cbc901234567890abcdef01234567" {
		t.Errorf("Data() saved To = %v, want [b4ebe3e2200cbc901234567890abcdef01234567]", savedEmail.To)
	}

	if savedEmail.Subject != "Test Subject" {
		t.Errorf("Data() saved Subject = %v, want %v", savedEmail.Subject, "Test Subject")
	}

	if savedEmail.Body != "Test email body" {
		t.Errorf("Data() saved Body = %v, want %v", savedEmail.Body, "Test email body")
	}

	if mockStore.savedTo[0] != "b4ebe3e2200cbc901234567890abcdef01234567" {
		t.Errorf("Data() saved to addressBox = %v, want %v", mockStore.savedTo[0], "b4ebe3e2200cbc901234567890abcdef01234567")
	}
}

func TestSession_Data_HTML(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{
		Store: mockStore,
		From:  "sender@example.com",
		To:    []string{"b4ebe3e2200cbc901234567890abcdef01234567"},
	}

	emailData := `From: sender@example.com
To: b4ebe3e2200cbc901234567890abcdef01234567@example.com
Subject: HTML Test
Content-Type: text/html; charset=UTF-8

<html><body><h1>HTML Content</h1></body></html>`

	err := session.Data(strings.NewReader(emailData))
	if err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	if len(mockStore.savedEmails) != 1 {
		t.Fatalf("Data() saved %d emails, want 1", len(mockStore.savedEmails))
	}

	savedEmail := mockStore.savedEmails[0]
	expectedBody := "<html><body><h1>HTML Content</h1></body></html>"
	if savedEmail.Body != expectedBody {
		t.Errorf("Data() saved Body = %v, want %v", savedEmail.Body, expectedBody)
	}
}

func TestSession_Data_MultiPart(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{
		Store: mockStore,
		From:  "sender@example.com",
		To:    []string{"b4ebe3e2200cbc901234567890abcdef01234567"},
	}

	emailData := `From: sender@example.com
To: b4ebe3e2200cbc901234567890abcdef01234567@example.com
Subject: Multipart Test
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="boundary123"

--boundary123
Content-Type: text/plain; charset=UTF-8

Plain text content

--boundary123
Content-Type: text/html; charset=UTF-8

<html><body><h1>HTML Content</h1></body></html>

--boundary123--`

	err := session.Data(strings.NewReader(emailData))
	if err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	if len(mockStore.savedEmails) != 1 {
		t.Fatalf("Data() saved %d emails, want 1", len(mockStore.savedEmails))
	}

	savedEmail := mockStore.savedEmails[0]
	expectedBody := "<html><body><h1>HTML Content</h1></body></html>"
	if savedEmail.Body != expectedBody {
		t.Errorf("Data() saved Body = %v, want %v", savedEmail.Body, expectedBody)
	}
}

func TestSession_Data_InvalidEmail(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{
		Store: mockStore,
		From:  "sender@example.com",
		To:    []string{"b4ebe3e2200cbc901234567890abcdef01234567"},
	}

	invalidEmailData := `Invalid email content without proper headers`

	err := session.Data(strings.NewReader(invalidEmailData))
	if err == nil {
		t.Error("Data() expected error for invalid email, got nil")
	}
}

func TestSession_Data_EmptyBody(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{
		Store: mockStore,
		From:  "sender@example.com",
		To:    []string{"b4ebe3e2200cbc901234567890abcdef01234567"},
	}

	emailData := `From: sender@example.com
To: b4ebe3e2200cbc901234567890abcdef01234567@example.com
Subject: Empty Body Test
Content-Type: text/plain; charset=UTF-8

`

	err := session.Data(strings.NewReader(emailData))
	if err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	if len(mockStore.savedEmails) != 1 {
		t.Fatalf("Data() saved %d emails, want 1", len(mockStore.savedEmails))
	}

	savedEmail := mockStore.savedEmails[0]
	if savedEmail.Body != "" {
		t.Errorf("Data() saved Body = %v, want empty", savedEmail.Body)
	}
}

func TestSession_Data_MultipleRecipients(t *testing.T) {
	mockStore := &mockStore{}
	session := &Session{
		Store: mockStore,
		From:  "sender@example.com",
		To:    []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "cccccccccccccccccccccccccccccccccccc"},
	}

	emailData := `From: sender@example.com
To: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@example.com, bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb@example.com, ccccccccccccccccccccccccccccccccccccc@example.com
Subject: Multi-recipient Test
Content-Type: text/plain; charset=UTF-8

Test body`

	err := session.Data(strings.NewReader(emailData))
	if err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	if len(mockStore.savedEmails) != 3 {
		t.Fatalf("Data() saved %d emails, want 3 (one per recipient)", len(mockStore.savedEmails))
	}

	expectedRecipients := []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "cccccccccccccccccccccccccccccccccccc"}
	for i, expected := range expectedRecipients {
		if mockStore.savedTo[i] != expected {
			t.Errorf("Data() savedTo[%d] = %v, want %v", i, mockStore.savedTo[i], expected)
		}
	}
}
