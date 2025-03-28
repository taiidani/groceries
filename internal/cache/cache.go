package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) (err error)
	Get(ctx context.Context, key string, value any) error
}

const (
	dbPrefix = "groceries:"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

func NewRedisCache(client *redis.Client) Cache {
	return &RedisStore{Client: client}
}

func NewClient(ctx context.Context) *redis.Client {
	host, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		log.Fatalf("REDIS_HOST env var not found")
	}

	// Determine the address, whether it be HOST:PORT or HOST & PORT
	var port string
	if host, port, ok = strings.Cut(host, ":"); !ok {
		if port, ok = os.LookupEnv("REDIS_PORT"); !ok {
			port = "4646"
		}
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	opts := &redis.Options{Addr: addr}

	// Determine if a username & password is set
	if user, ok := os.LookupEnv("REDIS_USER"); ok {
		opts.TLSConfig = &tls.Config{}
		opts.Username = user
	}
	if pass, ok := os.LookupEnv("REDIS_PASSWORD"); ok {
		opts.TLSConfig = &tls.Config{}
		opts.Password = pass
	}

	client := redis.NewClient(opts)
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if cmd := client.Ping(ctxTimeout); cmd.Err() != nil {
		log.Fatalf("Unable to connect to Redis backend at %s: %s", addr, cmd.Err())
	}
	slog.Info("Redis persistence configured", "addr", addr)

	return client
}
