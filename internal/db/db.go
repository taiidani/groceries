package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/taiidani/groceries/internal/data"
)

// db is a singleton holding either a Redis or Memory backed database
var db data.DB

func NewDB() data.DB {
	host, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		// Default to a memory backend
		slog.Warn("Redis persistence disabled")
		db = &data.MemoryStore{Data: map[string][]byte{}}
		return db
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

	// Set the singleton db value to the Redis backend
	db = &data.RedisStore{Client: redis.NewClient(opts)}
	if err := db.Set(context.Background(), "client", "groceries", time.Hour*24); err != nil {
		log.Fatalf("Unable to connect to Redis backend at %s: %s", addr, err)
	}
	slog.Info("Redis persistence configured", "addr", addr)
	return db
}
