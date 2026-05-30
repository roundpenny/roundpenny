package cache

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
	mock   bool
	data   map[string]string
}

func NewClient() *Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "redis:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	c := &Client{client: rdb}

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("redis not available, using mock cache", "error", err)
		c.mock = true
		c.data = make(map[string]string)
	}

	return c
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if c.mock {
		v, ok := c.data[key]
		if !ok {
			return "", redis.Nil
		}
		return v, nil
	}
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if c.mock {
		c.data[key] = value
		return nil
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	if c.mock {
		for _, k := range keys {
			delete(c.data, k)
		}
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	if c.mock {
		_, ok := c.data[key]
		return ok, nil
	}
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if c.mock {
		return nil
	}
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	if c.mock {
		val := c.data[key]
		var n int64
		for _, ch := range val {
			n = n*10 + int64(ch-'0')
		}
		n++
		c.data[key] = itoa(n)
		return n, nil
	}
	return c.client.Incr(ctx, key).Result()
}

func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	if c.mock {
		val := c.data[key]
		var n int64
		for _, ch := range val {
			n = n*10 + int64(ch-'0')
		}
		n += value
		c.data[key] = itoa(n)
		return n, nil
	}
	return c.client.IncrBy(ctx, key, value).Result()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c.mock {
		return -1, nil
	}
	return c.client.TTL(ctx, key).Result()
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
