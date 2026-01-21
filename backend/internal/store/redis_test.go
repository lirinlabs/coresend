package store

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	addr := "localhost:6379"
	store := NewStore(addr, "")

	if store == nil {
		t.Fatal("NewStore() returned nil")
	}

	if store.client == nil {
		t.Fatal("NewStore() client is nil")
	}

	if store.client.Options().Addr != addr {
		t.Errorf("NewStore() addr = %v, want %v", store.client.Options().Addr, addr)
	}
}

func TestNewStoreWithPassword(t *testing.T) {
	addr := "localhost:6379"
	password := "testpassword"
	store := NewStore(addr, password)

	if store == nil {
		t.Fatal("NewStore() returned nil")
	}

	if store.client.Options().Password != password {
		t.Errorf("NewStore() password not set correctly")
	}
}

func TestSaveEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	store := NewStore("localhost:6379", "")

	defer func() {
		if err := store.client.Close(); err != nil {
			t.Logf("Error closing Redis client: %v", err)
		}
	}()

	if err := store.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	email := Email{
		From:       "sender@example.com",
		To:         []string{"recipient@example.com"},
		Subject:    "Test Subject",
		Body:       "Test Body",
		ReceivedAt: time.Now().UTC(),
	}

	addressBox := "test@example.com"

	err := store.SaveEmail(ctx, addressBox, email)
	if err != nil {
		t.Fatalf("SaveEmail() error = %v", err)
	}

	key := "inbox:" + addressBox
	exists, err := store.client.Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("Redis Exists() error = %v", err)
	}

	if exists == 0 {
		t.Fatal("Email was not saved to Redis")
	}

	ttl, err := store.client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("Redis TTL() error = %v", err)
	}

	expectedTTL := 24 * time.Hour
	if ttl < expectedTTL-time.Minute || ttl > expectedTTL+time.Minute {
		t.Errorf("TTL = %v, want ~%v", ttl, expectedTTL)
	}

	store.client.Del(ctx, key)
}

func TestSaveEmail_MultipleEmails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	store := NewStore("localhost:6379", "")

	defer func() {
		if err := store.client.Close(); err != nil {
			t.Logf("Error closing Redis client: %v", err)
		}
	}()

	if err := store.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	addressBox := "test@example.com"

	for i := 0; i < 3; i++ {
		email := Email{
			From:       "sender@example.com",
			To:         []string{addressBox},
			Subject:    "Test Subject",
			Body:       "Test Body",
			ReceivedAt: time.Now().UTC(),
		}

		err := store.SaveEmail(ctx, addressBox, email)
		if err != nil {
			t.Fatalf("SaveEmail() iteration %d error = %v", i, err)
		}
	}

	key := "inbox:" + addressBox
	length, err := store.client.LLen(ctx, key).Result()
	if err != nil {
		t.Fatalf("Redis LLen() error = %v", err)
	}

	if length != 3 {
		t.Errorf("Expected 3 emails in list, got %d", length)
	}

	store.client.Del(ctx, key)
}

func TestSaveEmail_GeneratesID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	store := NewStore("localhost:6379", "")

	defer func() {
		if err := store.client.Close(); err != nil {
			t.Logf("Error closing Redis client: %v", err)
		}
	}()

	if err := store.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	addressBox := "b4ebe3e2200cbc90"

	email := Email{
		From:       "sender@example.com",
		To:         []string{addressBox},
		Subject:    "Test Subject",
		Body:       "Test Body",
		ReceivedAt: time.Now().UTC(),
	}

	err := store.SaveEmail(ctx, addressBox, email)
	if err != nil {
		t.Fatalf("SaveEmail() error = %v", err)
	}

	store.client.Del(ctx, "inbox:"+addressBox)
}

func TestGetEmails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	store := NewStore("localhost:6379", "")

	defer func() {
		if err := store.client.Close(); err != nil {
			t.Logf("Error closing Redis client: %v", err)
		}
	}()

	if err := store.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	addressBox := "b4ebe3e2200cbc90"
	key := "inbox:" + addressBox

	// Clean up before test
	store.client.Del(ctx, key)

	// Save test emails
	for i := 0; i < 3; i++ {
		email := Email{
			From:       "sender@example.com",
			To:         []string{addressBox},
			Subject:    fmt.Sprintf("Test Subject %d", i),
			Body:       fmt.Sprintf("Test Body %d", i),
			ReceivedAt: time.Now().UTC(),
		}
		if err := store.SaveEmail(ctx, addressBox, email); err != nil {
			t.Fatalf("SaveEmail() error = %v", err)
		}
	}

	// Retrieve emails
	emails, err := store.GetEmails(ctx, addressBox)
	if err != nil {
		t.Fatalf("GetEmails() error = %v", err)
	}

	if len(emails) != 3 {
		t.Errorf("GetEmails() returned %d emails, want 3", len(emails))
	}

	// Emails should be in reverse order (newest first due to LPUSH)
	if emails[0].Subject != "Test Subject 2" {
		t.Errorf("GetEmails() first email subject = %v, want 'Test Subject 2'", emails[0].Subject)
	}

	// Each email should have an ID
	for i, email := range emails {
		if email.ID == "" {
			t.Errorf("GetEmails() email[%d] has empty ID", i)
		}
	}

	store.client.Del(ctx, key)
}

func TestGetEmails_EmptyInbox(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	store := NewStore("localhost:6379", "")

	defer func() {
		if err := store.client.Close(); err != nil {
			t.Logf("Error closing Redis client: %v", err)
		}
	}()

	if err := store.Ping(ctx); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	emails, err := store.GetEmails(ctx, "nonexistent1234ab")
	if err != nil {
		t.Fatalf("GetEmails() error = %v", err)
	}

	if len(emails) != 0 {
		t.Errorf("GetEmails() returned %d emails for nonexistent inbox, want 0", len(emails))
	}
}
