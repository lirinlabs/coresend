package smtp

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	gosmtp "github.com/emersion/go-smtp"
	"github.com/fn-jakubkarp/coresend/internal/addr"
	"github.com/fn-jakubkarp/coresend/internal/store"
)

type Backend struct {
	Store store.EmailStore
}

func (bkd *Backend) NewSession(c *gosmtp.Conn) (gosmtp.Session, error) {
	return &Session{Store: bkd.Store}, nil
}

type Session struct {
	Store store.EmailStore
	From  string
	To    []string
}

func (s *Session) Mail(from string, opts *gosmtp.MailOptions) error {
	log.Printf("MAIL FROM: %s", from)
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
	log.Printf("RCPT TO: %s", to)

	localPart := extractLocalPart(to)
	if !addr.IsValid(localPart) {
		log.Printf("Rejected invalid address: %s", to)
		return &gosmtp.SMTPError{
			Code:         550,
			EnhancedCode: gosmtp.EnhancedCode{5, 1, 1},
			Message:      "Mailbox does not exist",
		}
	}

	s.To = append(s.To, strings.ToLower(localPart))
	return nil
}

func extractLocalPart(email string) string {
	if idx := strings.LastIndex(email, "@"); idx != -1 {
		return email[:idx]
	}
	return email
}

func (s *Session) Data(r io.Reader) error {
	mr, err := mail.CreateReader(r)
	if err != nil {
		return err
	}

	email := store.Email{
		From:       s.From,
		To:         s.To,
		ReceivedAt: time.Now(),
	}

	if subject, err := mr.Header.Subject(); err == nil {
		email.Subject = subject
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading email part: %v", err)
			break
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, err := h.ContentType()
			if err != nil {
				log.Printf("Error reading content type: %v", err)
				continue
			}

			body, err := io.ReadAll(p.Body)
			if err != nil {
				log.Printf("Error reading body: %v", err)
				continue
			}

			// Prefer HTML over plain text when both are present
			if contentType == "text/html" {
				email.Body = string(body)
			} else if contentType == "text/plain" && email.Body == "" {
				email.Body = string(body)
			}

		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			log.Printf("Skipping attachment: %s (not supported)", filename)
		}
	}

	// Save email to each recipient's inbox
	var lastErr error
	for _, recipient := range s.To {
		if err := s.Store.SaveEmail(context.Background(), recipient, email); err != nil {
			log.Printf("Error saving email for %s: %v", recipient, err)
			lastErr = err
		}
	}

	if lastErr != nil {
		return fmt.Errorf("failed to save email to one or more recipients: %w", lastErr)
	}

	log.Printf("Email saved to %d recipient(s)", len(s.To))
	return nil
}

func (s *Session) Reset() {
	s.From = ""
	s.To = nil
}

func (s *Session) Logout() error {
	return nil
}
