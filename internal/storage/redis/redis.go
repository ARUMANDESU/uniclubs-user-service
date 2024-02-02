package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-user-service/internal/storage"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
	"time"
)

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
		log.Fatal(err)
	}

	return &Storage{client: client}, err
}

func (s Storage) Create(ctx context.Context, sessionToken string, userID int64, duration time.Duration) error {
	const op = "storage.redis.Create"

	err := s.client.Set(ctx, sessionToken, userID, duration).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Storage) Get(ctx context.Context, sessionToken string) (int64, error) {
	const op = "storage.redis.Get"

	val, err := s.client.Get(ctx, sessionToken).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrSessionNotExists)
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
	const op = "storage.redis.Delete"
	err := s.client.Del(ctx, sessionToken).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
