package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Email struct {
	ID         string    `json:"id"`
	From       string    `json:"from"`
	To         []string  `json:"to"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	ReceivedAt time.Time `json:"received_at"`
}

type EmailStore interface {
	SaveEmail(ctx context.Context, addressBox string, email Email) error
	GetEmails(ctx context.Context, addressBox string) ([]Email, error)
	GetEmail(ctx context.Context, addressBox string, emailID string) (*Email, error)
	DeleteEmail(ctx context.Context, addressBox string, emailID string) error
	ClearInbox(ctx context.Context, addressBox string) (int64, error)
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)
	RegisterAddress(ctx context.Context, addressBox string, duration time.Duration) error
	IsAddressActive(ctx context.Context, addressBox string) (bool, error)
	Ping(ctx context.Context) error
	CheckAndStoreNonce(ctx context.Context, nonce string, ttl time.Duration) (bool, error)
}

type Store struct {
	client *redis.Client
}

func NewStore(addr, password string) *Store {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &Store{client: rdb}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *Store) SaveEmail(ctx context.Context, addressBox string, email Email) error {
	if email.ID == "" {
		email.ID = uuid.New().String()
	}

	data, err := json.Marshal(email)
	if err != nil {
		return err
	}

	zKey := fmt.Sprintf("inbox:%s", addressBox)
	hKey := fmt.Sprintf("emails:%s", addressBox)

	now := float64(time.Now().Unix())

	pipe := s.client.Pipeline()

	pipe.ZAdd(ctx, zKey, redis.Z{Score: now, Member: email.ID})

	pipe.HSet(ctx, hKey, email.ID, data)

	pipe.ZRemRangeByRank(ctx, zKey, 0, -101) // Keep 100 latest emails

	pipe.Expire(ctx, zKey, 24*time.Hour)
	pipe.Expire(ctx, hKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) GetEmails(ctx context.Context, addressBox string) ([]Email, error) {
	zKey := fmt.Sprintf("inbox:%s", addressBox)
	hKey := fmt.Sprintf("emails:%s", addressBox)

	// Fetches newest IDs from the Sorted Set
	ids, err := s.client.ZRevRange(ctx, zKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []Email{}, nil
	}

	// HMGet returns interface{} slice, so we must handle type assertion
	rawData, err := s.client.HMGet(ctx, hKey, ids...).Result()
	if err != nil {
		return nil, err
	}

	emails := make([]Email, 0, len(rawData))
	for i, item := range rawData {
		if item == nil {
			slog.Warn("Skipping nil email data", "index", i, "id", ids[i])
			continue
		}

		strData, ok := item.(string)
		if !ok {
			slog.Warn("Skipping email with invalid type", "index", i, "id", ids[i])
			continue
		}

		var email Email
		if err := json.Unmarshal([]byte(strData), &email); err != nil {
			slog.Warn("Skipping unmarshalable email", "index", i, "id", ids[i], "error", err)
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

func (s *Store) GetEmail(ctx context.Context, addressBox string, emailID string) (*Email, error) {
	hKey := fmt.Sprintf("emails:%s", addressBox)

	data, err := s.client.HGet(ctx, hKey, emailID).Result()
	if err == redis.Nil {
		return nil, nil // Email not found
	} else if err != nil {
		return nil, err
	}

	var email Email
	if err := json.Unmarshal([]byte(data), &email); err != nil {
		return nil, err
	}

	return &email, nil
}

func (s *Store) DeleteEmail(ctx context.Context, addressBox string, emailID string) error {
	zKey := fmt.Sprintf("inbox:%s", addressBox)
	hKey := fmt.Sprintf("emails:%s", addressBox)

	pipe := s.client.Pipeline()

	pipe.ZRem(ctx, zKey, emailID)
	pipe.HDel(ctx, hKey, emailID)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) ClearInbox(ctx context.Context, addressBox string) (int64, error) {
	zKey := fmt.Sprintf("inbox:%s", addressBox)
	hKey := fmt.Sprintf("emails:%s", addressBox)

	return s.client.Del(ctx, zKey, hKey).Result()
}

func (s *Store) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	key = fmt.Sprintf("ratelimit:%s", key)

	var count int64
	ok, err := s.client.SetNX(ctx, key, 1, window).Result()
	if err != nil {
		return false, 0, err
	}

	if ok {
		count = 1
	} else {
		count, err = s.client.Incr(ctx, key).Result()
		if err != nil {
			return false, 0, err
		}
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return count <= int64(limit), remaining, nil
}

func (s *Store) RegisterAddress(ctx context.Context, addressBox string, duration time.Duration) error {
	key := fmt.Sprintf("active_address:%s", addressBox)
	return s.client.Set(ctx, key, "1", duration).Err()
}

func (s *Store) IsAddressActive(ctx context.Context, addressBox string) (bool, error) {
	key := fmt.Sprintf("active_address:%s", addressBox)

	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func (s *Store) CheckAndStoreNonce(ctx context.Context, nonce string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("nonce:%s", nonce)
	return s.client.SetNX(ctx, key, "1", ttl).Result()
}
