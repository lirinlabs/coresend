package store

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) (*Store, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	s := NewStore(mr.Addr(), "")
	t.Cleanup(func() {
		_ = s.client.Close()
	})

	return s, mr
}

func assertTTLWithin(t *testing.T, ttl, max time.Duration) {
	t.Helper()

	if ttl <= 0 {
		t.Fatalf("ttl = %s, want > 0", ttl)
	}
	if ttl > max {
		t.Fatalf("ttl = %s, want <= %s", ttl, max)
	}
}

func TestNewStoreAndPing(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		s, _ := newTestStore(t)
		if s == nil {
			t.Fatalf("NewStore returned nil")
		}
		if err := s.Ping(context.Background()); err != nil {
			t.Fatalf("Ping() returned error: %v", err)
		}
	})

	t.Run("failure when redis unavailable", func(t *testing.T) {
		t.Parallel()

		s, mr := newTestStore(t)
		mr.Close()

		if err := s.Ping(context.Background()); err == nil {
			t.Fatalf("Ping() expected error when redis is closed")
		}
	})
}

func TestSaveEmail_AutoIDAndTTL(t *testing.T) {
	t.Parallel()

	s, _ := newTestStore(t)
	ctx := context.Background()
	address := "abc123"
	zKey := "inbox:" + address
	hKey := "emails:" + address

	email := Email{
		From:       "alice@example.com",
		To:         []string{"abc123@coresend.io"},
		Subject:    "Hello",
		Body:       "World",
		ReceivedAt: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
	}

	if err := s.SaveEmail(ctx, address, email); err != nil {
		t.Fatalf("SaveEmail() error: %v", err)
	}

	ids, err := s.client.ZRevRange(ctx, zKey, 0, -1).Result()
	if err != nil {
		t.Fatalf("ZRevRange() error: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("zset ids length = %d, want %d", len(ids), 1)
	}

	generatedID := ids[0]
	if _, err := uuid.Parse(generatedID); err != nil {
		t.Fatalf("generated email id is not UUID: %q (%v)", generatedID, err)
	}

	rawEmail, err := s.client.HGet(ctx, hKey, generatedID).Result()
	if err != nil {
		t.Fatalf("HGet() error: %v", err)
	}
	var stored Email
	if err := json.Unmarshal([]byte(rawEmail), &stored); err != nil {
		t.Fatalf("failed to unmarshal stored email: %v", err)
	}
	if stored.Subject != email.Subject {
		t.Fatalf("stored subject = %q, want %q", stored.Subject, email.Subject)
	}

	ttlZ, err := s.client.TTL(ctx, zKey).Result()
	if err != nil {
		t.Fatalf("TTL(zKey) error: %v", err)
	}
	assertTTLWithin(t, ttlZ, 24*time.Hour)

	ttlH, err := s.client.TTL(ctx, hKey).Result()
	if err != nil {
		t.Fatalf("TTL(hKey) error: %v", err)
	}
	assertTTLWithin(t, ttlH, 24*time.Hour)
}

func TestSaveEmail_Enforces100Newest(t *testing.T) {
	t.Parallel()

	s, _ := newTestStore(t)
	ctx := context.Background()
	address := "retention"

	for i := 0; i < 101; i++ {
		err := s.SaveEmail(ctx, address, Email{
			ID:         fmt.Sprintf("id-%03d", i),
			From:       "sender@example.com",
			To:         []string{"retention@coresend.io"},
			Subject:    fmt.Sprintf("subject-%03d", i),
			Body:       "body",
			ReceivedAt: time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("SaveEmail(%d) error: %v", i, err)
		}
	}

	count, err := s.client.ZCard(ctx, "inbox:"+address).Result()
	if err != nil {
		t.Fatalf("ZCard() error: %v", err)
	}
	if count != 100 {
		t.Fatalf("zset count = %d, want %d", count, 100)
	}

	emails, err := s.GetEmails(ctx, address)
	if err != nil {
		t.Fatalf("GetEmails() error: %v", err)
	}
	if len(emails) != 100 {
		t.Fatalf("GetEmails length = %d, want %d", len(emails), 100)
	}
}

func TestGetEmails(t *testing.T) {
	t.Parallel()

	t.Run("empty inbox", func(t *testing.T) {
		t.Parallel()

		s, _ := newTestStore(t)
		emails, err := s.GetEmails(context.Background(), "empty")
		if err != nil {
			t.Fatalf("GetEmails() error: %v", err)
		}
		if len(emails) != 0 {
			t.Fatalf("emails length = %d, want %d", len(emails), 0)
		}
	})

	t.Run("returns newest first and skips missing or invalid entries", func(t *testing.T) {
		t.Parallel()

		s, _ := newTestStore(t)
		ctx := context.Background()
		address := "ordered"
		zKey := "inbox:" + address
		hKey := "emails:" + address

		validA := Email{ID: "id-new", Subject: "new", ReceivedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)}
		validB := Email{ID: "id-mid", Subject: "mid", ReceivedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)}
		validC := Email{ID: "id-old", Subject: "old", ReceivedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}

		rawA, _ := json.Marshal(validA)
		rawB, _ := json.Marshal(validB)
		rawC, _ := json.Marshal(validC)

		if err := s.client.HSet(ctx, hKey,
			validA.ID, string(rawA),
			validB.ID, string(rawB),
			validC.ID, string(rawC),
			"id-bad", "not-json",
		).Err(); err != nil {
			t.Fatalf("HSet() error: %v", err)
		}

		if err := s.client.ZAdd(ctx, zKey,
			redis.Z{Score: 50, Member: validA.ID},
			redis.Z{Score: 40, Member: validB.ID},
			redis.Z{Score: 30, Member: validC.ID},
			redis.Z{Score: 20, Member: "id-missing"},
			redis.Z{Score: 10, Member: "id-bad"},
		).Err(); err != nil {
			t.Fatalf("ZAdd() error: %v", err)
		}

		emails, err := s.GetEmails(ctx, address)
		if err != nil {
			t.Fatalf("GetEmails() error: %v", err)
		}
		if len(emails) != 3 {
			t.Fatalf("emails length = %d, want %d", len(emails), 3)
		}
		if emails[0].ID != "id-new" || emails[1].ID != "id-mid" || emails[2].ID != "id-old" {
			t.Fatalf("unexpected email order: [%s %s %s]", emails[0].ID, emails[1].ID, emails[2].ID)
		}
	})
}

func TestGetEmail(t *testing.T) {
	t.Parallel()

	t.Run("returns existing email", func(t *testing.T) {
		t.Parallel()

		s, _ := newTestStore(t)
		ctx := context.Background()
		address := "one"
		id := "email-1"
		email := Email{ID: id, Subject: "hello", ReceivedAt: time.Now().UTC()}
		raw, _ := json.Marshal(email)

		if err := s.client.HSet(ctx, "emails:"+address, id, string(raw)).Err(); err != nil {
			t.Fatalf("HSet() error: %v", err)
		}

		got, err := s.GetEmail(ctx, address, id)
		if err != nil {
			t.Fatalf("GetEmail() error: %v", err)
		}
		if got == nil {
			t.Fatalf("GetEmail() returned nil email")
		}
		if got.ID != id {
			t.Fatalf("email id = %q, want %q", got.ID, id)
		}
	})

	t.Run("returns nil when missing", func(t *testing.T) {
		t.Parallel()

		s, _ := newTestStore(t)
		got, err := s.GetEmail(context.Background(), "missing", "missing-id")
		if err != nil {
			t.Fatalf("GetEmail() error: %v", err)
		}
		if got != nil {
			t.Fatalf("GetEmail() = %#v, want nil", got)
		}
	})

	t.Run("returns error for invalid json", func(t *testing.T) {
		t.Parallel()

		s, _ := newTestStore(t)
		ctx := context.Background()
		if err := s.client.HSet(ctx, "emails:bad", "id-1", "not-json").Err(); err != nil {
			t.Fatalf("HSet() error: %v", err)
		}

		_, err := s.GetEmail(ctx, "bad", "id-1")
		if err == nil {
			t.Fatalf("GetEmail() expected unmarshal error")
		}
	})
}

func TestDeleteEmail(t *testing.T) {
	t.Parallel()

	s, _ := newTestStore(t)
	ctx := context.Background()
	address := "delete-me"
	id := "id-1"
	zKey := "inbox:" + address
	hKey := "emails:" + address

	if err := s.client.ZAdd(ctx, zKey, redis.Z{Score: 1, Member: id}).Err(); err != nil {
		t.Fatalf("ZAdd() error: %v", err)
	}
	if err := s.client.HSet(ctx, hKey, id, `{"id":"id-1"}`).Err(); err != nil {
		t.Fatalf("HSet() error: %v", err)
	}

	if err := s.DeleteEmail(ctx, address, id); err != nil {
		t.Fatalf("DeleteEmail() error: %v", err)
	}

	zCount, err := s.client.ZCard(ctx, zKey).Result()
	if err != nil {
		t.Fatalf("ZCard() error: %v", err)
	}
	if zCount != 0 {
		t.Fatalf("zset count = %d, want %d", zCount, 0)
	}

	exists, err := s.client.HExists(ctx, hKey, id).Result()
	if err != nil {
		t.Fatalf("HExists() error: %v", err)
	}
	if exists {
		t.Fatalf("hash field %q still exists", id)
	}
}

func TestClearInbox(t *testing.T) {
	t.Parallel()

	s, _ := newTestStore(t)
	ctx := context.Background()
	address := "clear-me"
	zKey := "inbox:" + address
	hKey := "emails:" + address

	if err := s.client.ZAdd(ctx, zKey, redis.Z{Score: 1, Member: "id-1"}).Err(); err != nil {
		t.Fatalf("ZAdd() error: %v", err)
	}
	if err := s.client.HSet(ctx, hKey, "id-1", `{"id":"id-1"}`).Err(); err != nil {
		t.Fatalf("HSet() error: %v", err)
	}

	deleted, err := s.ClearInbox(ctx, address)
	if err != nil {
		t.Fatalf("ClearInbox() error: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("deleted key count = %d, want %d", deleted, 2)
	}

	if exists, _ := s.client.Exists(ctx, zKey).Result(); exists != 0 {
		t.Fatalf("zset key still exists")
	}
	if exists, _ := s.client.Exists(ctx, hKey).Result(); exists != 0 {
		t.Fatalf("hash key still exists")
	}
}

func TestCheckRateLimit(t *testing.T) {
	t.Parallel()

	s, _ := newTestStore(t)
	ctx := context.Background()
	window := time.Minute
	key := "inbox:192.0.2.10:1234"
	limit := 2

	allowed, remaining, err := s.CheckRateLimit(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("CheckRateLimit() first call error: %v", err)
	}
	if !allowed || remaining != 1 {
		t.Fatalf("first call => allowed=%v remaining=%d, want true/1", allowed, remaining)
	}

	allowed, remaining, err = s.CheckRateLimit(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("CheckRateLimit() second call error: %v", err)
	}
	if !allowed || remaining != 0 {
		t.Fatalf("second call => allowed=%v remaining=%d, want true/0", allowed, remaining)
	}

	allowed, remaining, err = s.CheckRateLimit(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("CheckRateLimit() third call error: %v", err)
	}
	if allowed || remaining != 0 {
		t.Fatalf("third call => allowed=%v remaining=%d, want false/0", allowed, remaining)
	}

	redisKey := "ratelimit:" + key
	ttl, err := s.client.TTL(ctx, redisKey).Result()
	if err != nil {
		t.Fatalf("TTL() error: %v", err)
	}
	assertTTLWithin(t, ttl, window)

	value, err := s.client.Get(ctx, redisKey).Int64()
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if value != 3 {
		t.Fatalf("rate limit counter = %d, want %d", value, 3)
	}
}

func TestRegisterAddressAndIsAddressActive(t *testing.T) {
	t.Parallel()

	s, mr := newTestStore(t)
	ctx := context.Background()
	address := "active-box"
	ttlWindow := 2 * time.Minute

	active, err := s.IsAddressActive(ctx, address)
	if err != nil {
		t.Fatalf("IsAddressActive() pre-check error: %v", err)
	}
	if active {
		t.Fatalf("address should be inactive before registration")
	}

	if err := s.RegisterAddress(ctx, address, ttlWindow); err != nil {
		t.Fatalf("RegisterAddress() error: %v", err)
	}

	active, err = s.IsAddressActive(ctx, address)
	if err != nil {
		t.Fatalf("IsAddressActive() post-register error: %v", err)
	}
	if !active {
		t.Fatalf("address should be active after registration")
	}

	ttl, err := s.client.TTL(ctx, "active_address:"+address).Result()
	if err != nil {
		t.Fatalf("TTL() error: %v", err)
	}
	assertTTLWithin(t, ttl, ttlWindow)

	mr.FastForward(3 * time.Minute)
	active, err = s.IsAddressActive(ctx, address)
	if err != nil {
		t.Fatalf("IsAddressActive() after expiry error: %v", err)
	}
	if active {
		t.Fatalf("address should be inactive after ttl expiry")
	}
}

func TestCheckAndStoreNonce(t *testing.T) {
	t.Parallel()

	s, _ := newTestStore(t)
	ctx := context.Background()
	nonce := "550e8400-e29b-41d4-a716-446655440000"
	ttlWindow := 5 * time.Minute

	unique, err := s.CheckAndStoreNonce(ctx, nonce, ttlWindow)
	if err != nil {
		t.Fatalf("CheckAndStoreNonce() first call error: %v", err)
	}
	if !unique {
		t.Fatalf("first nonce insert should be unique")
	}

	unique, err = s.CheckAndStoreNonce(ctx, nonce, ttlWindow)
	if err != nil {
		t.Fatalf("CheckAndStoreNonce() second call error: %v", err)
	}
	if unique {
		t.Fatalf("second nonce insert should not be unique")
	}

	ttl, err := s.client.TTL(ctx, "nonce:"+nonce).Result()
	if err != nil {
		t.Fatalf("TTL() error: %v", err)
	}
	assertTTLWithin(t, ttl, ttlWindow)
}
