package store

import (
	"context"
	"encoding/json"
	"fmt"
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
	Ping(ctx context.Context) error
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

	key := fmt.Sprintf("inbox:%s", addressBox)

	// TODO: TTL resets for entire inbox on each new email, not per-email.
	// Consider ZSET with timestamp scores for per-email expiration.
	pipe := s.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, 99)
	pipe.Expire(ctx, key, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) GetEmails(ctx context.Context, addressBox string) ([]Email, error) {
	key := fmt.Sprintf("inbox:%s", addressBox)

	data, err := s.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	emails := make([]Email, 0, len(data))
	for _, item := range data {
		var email Email
		if err := json.Unmarshal([]byte(item), &email); err != nil {
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}
