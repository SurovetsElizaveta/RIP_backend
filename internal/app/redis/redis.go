package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"rip/internal/app/config"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	client *redis.Client
}

func New(cfg config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

func (c *Client) AddToBlacklist(ctx context.Context, token string, expiresIn time.Duration) error {
	key := "jwt_blacklist:" + token
	return c.Set(ctx, key, "1", expiresIn)
}

func (c *Client) IsInBlacklist(ctx context.Context, token string) (bool, error) {
	key := "jwt_blacklist:" + token
	result, err := c.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (c *Client) StoreRefreshToken(ctx context.Context, userID uint, token string, expiresIn time.Duration) error {
	key := "refresh_token:" + strconv.FormatUint(uint64(userID), 10)
	return c.Set(ctx, key, token, expiresIn)
}

func (c *Client) GetRefreshToken(ctx context.Context, userID uint) (string, error) {
	key := "refresh_token:" + strconv.FormatUint(uint64(userID), 10)
	return c.Get(ctx, key)
}

func (c *Client) DeleteRefreshToken(ctx context.Context, userID uint) error {
	key := "refresh_token:" + strconv.FormatUint(uint64(userID), 10)
	return c.Del(ctx, key)
}
