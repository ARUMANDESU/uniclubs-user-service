package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
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

func (s Storage) Create(ctx context.Context, sessionToken string, userID int64) error {
	const op = "storage.redis.Create"

	err := s.client.Set(ctx, sessionToken, userID, time.Hour)
	if err.Err() != nil {
		return fmt.Errorf("%s: %w", op, err.Err())
	}

	return nil
}

func (s Storage) Get(ctx context.Context, sessionToken string) (userID int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) Delete(ctx context.Context, sessionToken string) error {
	const op = "storage.redis.Delete"
	err := s.client.Del(ctx, sessionToken)
	if err.Err() != nil {
		return fmt.Errorf("%s: %w", op, err.Err())
	}

	return nil
}
