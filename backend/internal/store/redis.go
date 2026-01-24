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
	GetEmail(ctx context.Context, addressBox string, emailID string) (*Email, error)
	DeleteEmail(ctx context.Context, addressBox string, emailID string) error
	ClearInbox(ctx context.Context, addressBox string) (int64, error)
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error)
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

	now := float64(time.Now().Unix())
	pipe := s.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: now, Member: data})
	pipe.ZRemRangeByRank(ctx, key, 100, -1)
	pipe.Expire(ctx, key, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (s *Store) GetEmails(ctx context.Context, addressBox string) ([]Email, error) {
	key := fmt.Sprintf("inbox:%s", addressBox)

	data, err := s.client.ZRevRange(ctx, key, 0, -1).Result()
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

func (s *Store) GetEmail(ctx context.Context, addressBox string, emailID string) (*Email, error) {
	emails, err := s.GetEmails(ctx, addressBox)
	if err != nil {
		return nil, err
	}

	for _, email := range emails {
		if email.ID == emailID {
			return &email, nil
		}
	}

	return nil, nil
}

func (s *Store) DeleteEmail(ctx context.Context, addressBox string, emailID string) error {
	key := fmt.Sprintf("inbox:%s", addressBox)

	emails, err := s.GetEmails(ctx, addressBox)
	if err != nil {
		return err
	}

	for _, email := range emails {
		if email.ID == emailID {
			data, err := json.Marshal(email)
			if err != nil {
				return err
			}
			return s.client.ZRem(ctx, key, data).Err()
		}
	}

	return nil
}

func (s *Store) ClearInbox(ctx context.Context, addressBox string) (int64, error) {
	key := fmt.Sprintf("inbox:%s", addressBox)
	return s.client.Del(ctx, key).Result()
}

func (s *Store) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	now := time.Now()
	key = fmt.Sprintf("ratelimit:%s", key)

	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.ExpireAt(ctx, key, now.Add(window))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := int(incr.Val())
	return count <= limit, limit - count, nil
}
