package data

import (
	"context"
	"errors"
	"time"
)

type DB interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) (err error)
	Get(ctx context.Context, key string, value any) error
}

const (
	dbPrefix = "groceries:"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)
