package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const rateLimitTimeFormat = time.RFC3339

type Storage struct {
	client *redis.Client
}

func New(redisURL string) (*Storage, error) {
	const op = "storage.redis.New"

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := redis.NewClient(opt)

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to ping:   %w", op, err)
	}

	return &Storage{client: client}, err
}

func (s Storage) Close() error {
	const op = "storage.redis.close"

	err := s.client.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

func (s Storage) Create(ctx context.Context, sessionToken string, userID int64, duration time.Duration) error {
	const op = "storage.redis.create"

	err := s.client.Set(ctx, sessionToken, userID, duration).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Storage) Get(ctx context.Context, sessionToken string) (int64, error) {
	const op = "storage.redis.get"

	val, err := s.client.Get(ctx, sessionToken).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, fmt.Errorf("%s: %w", op, domain.ErrTokenNotFound)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	userID, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return int64(userID), nil
}

func (s Storage) Delete(ctx context.Context, sessionToken string) error {
	const op = "storage.redis.delete"
	err := s.client.Del(ctx, sessionToken).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Storage) GetLastRequestTime(ctx context.Context, email string) (time.Time, error) {

	val, err := s.client.Get(ctx, email).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return time.Time{}, domain.ErrNotFound
		}
		return time.Time{}, err
	}

	t, err := time.Parse(rateLimitTimeFormat, val)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func (s Storage) UpdateLastRequestTime(ctx context.Context, email string, t time.Time) error {
	return s.client.Set(ctx, email, t.Format(rateLimitTimeFormat), 0).Err()
}
