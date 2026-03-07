package smtp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/fn-jakubkarp/coresend/internal/store"
)

const smtpValidHexAddress = "0123456789abcdef0123456789abcdef01234567"

type smtpFakeStore struct {
	saveEmailFn          func(ctx context.Context, addressBox string, email store.Email) error
	getEmailsFn          func(ctx context.Context, addressBox string) ([]store.Email, error)
	getEmailFn           func(ctx context.Context, addressBox string, emailID string) (*store.Email, error)
	deleteEmailFn        func(ctx context.Context, addressBox string, emailID string) error
	clearInboxFn         func(ctx context.Context, addressBox string) (int64, error)
	checkRateLimitFn     func(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)
	registerAddressFn    func(ctx context.Context, addressBox string, duration time.Duration) error
	isAddressActiveFn    func(ctx context.Context, addressBox string) (bool, error)
	pingFn               func(ctx context.Context) error
	checkAndStoreNonceFn func(ctx context.Context, nonce string, ttl time.Duration) (bool, error)

	saveCalls      []smtpSaveCall
	isActiveCalls  []string
	pingErr        error
	registerErr    error
	checkNonceResp bool
	checkNonceSet  bool
}

type smtpSaveCall struct {
	addressBox string
	email      store.Email
}

func (f *smtpFakeStore) SaveEmail(ctx context.Context, addressBox string, email store.Email) error {
	f.saveCalls = append(f.saveCalls, smtpSaveCall{addressBox: addressBox, email: email})
	if f.saveEmailFn != nil {
		return f.saveEmailFn(ctx, addressBox, email)
	}
	return nil
}

func (f *smtpFakeStore) GetEmails(ctx context.Context, addressBox string) ([]store.Email, error) {
	if f.getEmailsFn != nil {
		return f.getEmailsFn(ctx, addressBox)
	}
	panic(fmt.Sprintf("unexpected GetEmails call: addressBox=%q", addressBox))
}

func (f *smtpFakeStore) GetEmail(ctx context.Context, addressBox string, emailID string) (*store.Email, error) {
	if f.getEmailFn != nil {
		return f.getEmailFn(ctx, addressBox, emailID)
	}
	panic(fmt.Sprintf("unexpected GetEmail call: addressBox=%q emailID=%q", addressBox, emailID))
}

func (f *smtpFakeStore) DeleteEmail(ctx context.Context, addressBox string, emailID string) error {
	if f.deleteEmailFn != nil {
		return f.deleteEmailFn(ctx, addressBox, emailID)
	}
	panic(fmt.Sprintf("unexpected DeleteEmail call: addressBox=%q emailID=%q", addressBox, emailID))
}

func (f *smtpFakeStore) ClearInbox(ctx context.Context, addressBox string) (int64, error) {
	if f.clearInboxFn != nil {
		return f.clearInboxFn(ctx, addressBox)
	}
	panic(fmt.Sprintf("unexpected ClearInbox call: addressBox=%q", addressBox))
}

func (f *smtpFakeStore) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	if f.checkRateLimitFn != nil {
		return f.checkRateLimitFn(ctx, key, limit, window)
	}
	panic(fmt.Sprintf("unexpected CheckRateLimit call: key=%q limit=%d window=%s", key, limit, window))
}

func (f *smtpFakeStore) RegisterAddress(ctx context.Context, addressBox string, duration time.Duration) error {
	if f.registerAddressFn != nil {
		return f.registerAddressFn(ctx, addressBox, duration)
	}
	if f.registerErr != nil {
		return f.registerErr
	}
	panic(fmt.Sprintf("unexpected RegisterAddress call: addressBox=%q duration=%s", addressBox, duration))
}

func (f *smtpFakeStore) IsAddressActive(ctx context.Context, addressBox string) (bool, error) {
	f.isActiveCalls = append(f.isActiveCalls, addressBox)
	if f.isAddressActiveFn != nil {
		return f.isAddressActiveFn(ctx, addressBox)
	}
	panic(fmt.Sprintf("unexpected IsAddressActive call: addressBox=%q", addressBox))
}

func (f *smtpFakeStore) Ping(ctx context.Context) error {
	if f.pingFn != nil {
		return f.pingFn(ctx)
	}
	if f.pingErr != nil {
		return f.pingErr
	}
	panic("unexpected Ping call")
}

func (f *smtpFakeStore) CheckAndStoreNonce(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
	if f.checkAndStoreNonceFn != nil {
		return f.checkAndStoreNonceFn(ctx, nonce, ttl)
	}
	if f.checkNonceSet {
		return f.checkNonceResp, nil
	}
	panic(fmt.Sprintf("unexpected CheckAndStoreNonce call: nonce=%q ttl=%s", nonce, ttl))
}

func requireSMTPErrorCode(t *testing.T, err error, wantCode int) {
	t.Helper()

	var smtpErr *gosmtp.SMTPError
	if !errors.As(err, &smtpErr) {
		t.Fatalf("expected *gosmtp.SMTPError, got %T (%v)", err, err)
	}
	if smtpErr.Code != wantCode {
		t.Fatalf("smtp error code = %d, want %d", smtpErr.Code, wantCode)
	}
}

type forcedReadErrorReader struct{}

func (forcedReadErrorReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("forced read error")
}

func TestBackend_NewSession(t *testing.T) {
	t.Parallel()

	s := &smtpFakeStore{}
	backend := &Backend{Store: s}

	gotSession, err := backend.NewSession(nil)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	session, ok := gotSession.(*Session)
	if !ok {
		t.Fatalf("session type = %T, want *Session", gotSession)
	}
	if session.Store != s {
		t.Fatalf("session store was not propagated")
	}

	if err := session.Logout(); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
}

func TestSession_Mail(t *testing.T) {
	t.Parallel()

	session := &Session{}
	if err := session.Mail("alice@example.com", nil); err != nil {
		t.Fatalf("Mail() error = %v", err)
	}
	if session.From != "alice@example.com" {
		t.Fatalf("From = %q, want %q", session.From, "alice@example.com")
	}
}

func TestSession_Rcpt(t *testing.T) {
	t.Parallel()

	t.Run("malformed address returns 550", func(t *testing.T) {
		t.Parallel()

		session := &Session{Store: &smtpFakeStore{}}
		err := session.Rcpt("not-a-hex@example.com", nil)
		if err == nil {
			t.Fatalf("Rcpt() expected error")
		}
		requireSMTPErrorCode(t, err, 550)
	})

	t.Run("store error returns 451", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{
			isAddressActiveFn: func(ctx context.Context, addressBox string) (bool, error) {
				return false, fmt.Errorf("redis down")
			},
		}
		session := &Session{Store: fakeStore}

		err := session.Rcpt(strings.ToUpper(smtpValidHexAddress)+"@example.com", nil)
		if err == nil {
			t.Fatalf("Rcpt() expected error")
		}
		requireSMTPErrorCode(t, err, 451)

		if len(fakeStore.isActiveCalls) != 1 {
			t.Fatalf("isAddressActive call count = %d, want 1", len(fakeStore.isActiveCalls))
		}
		if fakeStore.isActiveCalls[0] != smtpValidHexAddress {
			t.Fatalf("isAddressActive address = %q, want %q", fakeStore.isActiveCalls[0], smtpValidHexAddress)
		}
	})

	t.Run("inactive address returns 550", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{
			isAddressActiveFn: func(ctx context.Context, addressBox string) (bool, error) {
				return false, nil
			},
		}
		session := &Session{Store: fakeStore}

		err := session.Rcpt(smtpValidHexAddress+"@example.com", nil)
		if err == nil {
			t.Fatalf("Rcpt() expected error")
		}
		requireSMTPErrorCode(t, err, 550)
	})

	t.Run("active address is accepted and lowercased", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{
			isAddressActiveFn: func(ctx context.Context, addressBox string) (bool, error) {
				return true, nil
			},
		}
		session := &Session{Store: fakeStore}

		err := session.Rcpt(strings.ToUpper(smtpValidHexAddress)+"@example.com", nil)
		if err != nil {
			t.Fatalf("Rcpt() error = %v", err)
		}
		if len(session.To) != 1 {
			t.Fatalf("recipient count = %d, want 1", len(session.To))
		}
		if session.To[0] != smtpValidHexAddress {
			t.Fatalf("recipient = %q, want %q", session.To[0], smtpValidHexAddress)
		}
	})
}

func plainMessage(subject, body string) string {
	return fmt.Sprintf(
		"From: sender@example.com\r\n"+
			"To: recipient@example.com\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"\r\n"+
			"%s",
		subject,
		body,
	)
}

func multipartAlternativeMessage(subject, plainBody, htmlBody string) string {
	return fmt.Sprintf(
		"From: sender@example.com\r\n"+
			"To: recipient@example.com\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/alternative; boundary=ALT-BOUNDARY\r\n"+
			"\r\n"+
			"--ALT-BOUNDARY\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"\r\n"+
			"%s\r\n"+
			"--ALT-BOUNDARY\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"\r\n"+
			"%s\r\n"+
			"--ALT-BOUNDARY--\r\n",
		subject,
		plainBody,
		htmlBody,
	)
}

func multipartWithAttachmentMessage(subject, plainBody string) string {
	return fmt.Sprintf(
		"From: sender@example.com\r\n"+
			"To: recipient@example.com\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/mixed; boundary=MIXED-BOUNDARY\r\n"+
			"\r\n"+
			"--MIXED-BOUNDARY\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n"+
			"\r\n"+
			"%s\r\n"+
			"--MIXED-BOUNDARY\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"Content-Disposition: attachment; filename=\"file.txt\"\r\n"+
			"\r\n"+
			"attachment-contents\r\n"+
			"--MIXED-BOUNDARY--\r\n",
		subject,
		plainBody,
	)
}

func TestSession_Data(t *testing.T) {
	t.Parallel()

	t.Run("invalid reader returns parse error", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{}
		session := &Session{
			Store: fakeStore,
			From:  "sender@example.com",
			To:    []string{"recipient-a"},
		}

		err := session.Data(forcedReadErrorReader{})
		if err == nil {
			t.Fatalf("Data() expected error")
		}
		if len(fakeStore.saveCalls) != 0 {
			t.Fatalf("save call count = %d, want 0", len(fakeStore.saveCalls))
		}
	})

	t.Run("plain text body extraction", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{}
		session := &Session{
			Store: fakeStore,
			From:  "sender@example.com",
			To:    []string{"recipient-a"},
		}

		err := session.Data(strings.NewReader(plainMessage("Plain Subject", "Hello plain body")))
		if err != nil {
			t.Fatalf("Data() error = %v", err)
		}

		if len(fakeStore.saveCalls) != 1 {
			t.Fatalf("save call count = %d, want 1", len(fakeStore.saveCalls))
		}
		saved := fakeStore.saveCalls[0]
		if saved.addressBox != "recipient-a" {
			t.Fatalf("saved address = %q, want %q", saved.addressBox, "recipient-a")
		}
		if saved.email.Subject != "Plain Subject" {
			t.Fatalf("saved subject = %q, want %q", saved.email.Subject, "Plain Subject")
		}
		if saved.email.Body != "Hello plain body" {
			t.Fatalf("saved body = %q, want %q", saved.email.Body, "Hello plain body")
		}
		if saved.email.From != "sender@example.com" {
			t.Fatalf("saved from = %q, want %q", saved.email.From, "sender@example.com")
		}
		if len(saved.email.To) != 1 || saved.email.To[0] != "recipient-a" {
			t.Fatalf("saved recipients = %v, want [recipient-a]", saved.email.To)
		}
		if saved.email.ReceivedAt.IsZero() {
			t.Fatalf("saved receivedAt is zero")
		}
	})

	t.Run("html preferred over plain text", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{}
		session := &Session{
			Store: fakeStore,
			From:  "sender@example.com",
			To:    []string{"recipient-a"},
		}

		err := session.Data(strings.NewReader(multipartAlternativeMessage("Subject", "plain-body", "<b>html-body</b>")))
		if err != nil {
			t.Fatalf("Data() error = %v", err)
		}

		if len(fakeStore.saveCalls) != 1 {
			t.Fatalf("save call count = %d, want 1", len(fakeStore.saveCalls))
		}
		if got := fakeStore.saveCalls[0].email.Body; got != "<b>html-body</b>" {
			t.Fatalf("saved body = %q, want %q", got, "<b>html-body</b>")
		}
	})

	t.Run("attachment part is ignored", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{}
		session := &Session{
			Store: fakeStore,
			From:  "sender@example.com",
			To:    []string{"recipient-a"},
		}

		err := session.Data(strings.NewReader(multipartWithAttachmentMessage("Attachment Subject", "body-before-attachment")))
		if err != nil {
			t.Fatalf("Data() error = %v", err)
		}

		if len(fakeStore.saveCalls) != 1 {
			t.Fatalf("save call count = %d, want 1", len(fakeStore.saveCalls))
		}
		if got := fakeStore.saveCalls[0].email.Body; got != "body-before-attachment" {
			t.Fatalf("saved body = %q, want %q", got, "body-before-attachment")
		}
	})

	t.Run("save error for one recipient returns error", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{
			saveEmailFn: func(ctx context.Context, addressBox string, email store.Email) error {
				if addressBox == "recipient-b" {
					return fmt.Errorf("save failed")
				}
				return nil
			},
		}
		session := &Session{
			Store: fakeStore,
			From:  "sender@example.com",
			To:    []string{"recipient-a", "recipient-b"},
		}

		err := session.Data(strings.NewReader(plainMessage("Subject", "Body")))
		if err == nil {
			t.Fatalf("Data() expected error")
		}
		if !strings.Contains(err.Error(), "failed to save email to one or more recipients") {
			t.Fatalf("error = %q, expected storage failure message", err.Error())
		}
		if len(fakeStore.saveCalls) != 2 {
			t.Fatalf("save call count = %d, want 2", len(fakeStore.saveCalls))
		}
	})

	t.Run("success saves email to all recipients", func(t *testing.T) {
		t.Parallel()

		fakeStore := &smtpFakeStore{}
		session := &Session{
			Store: fakeStore,
			From:  "sender@example.com",
			To:    []string{"recipient-a", "recipient-b"},
		}

		err := session.Data(strings.NewReader(plainMessage("Subject", "Body")))
		if err != nil {
			t.Fatalf("Data() error = %v", err)
		}
		if len(fakeStore.saveCalls) != 2 {
			t.Fatalf("save call count = %d, want 2", len(fakeStore.saveCalls))
		}
	})
}

func TestSession_ResetAndLogout(t *testing.T) {
	t.Parallel()

	session := &Session{
		From: "sender@example.com",
		To:   []string{"recipient-a", "recipient-b"},
	}

	session.Reset()
	if session.From != "" {
		t.Fatalf("From after reset = %q, want empty", session.From)
	}
	if len(session.To) != 0 {
		t.Fatalf("To after reset = %v, want nil", session.To)
	}

	if err := session.Logout(); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
}
