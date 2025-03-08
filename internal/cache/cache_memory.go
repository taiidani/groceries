package cache

import (
	"context"
	"encoding/json"
	"time"
)

type MemoryStore struct {
	Data map[string][]byte
}

func (s *MemoryStore) Set(ctx context.Context, key string, value any, expiration time.Duration) (err error) {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	s.Data[dbPrefix+key] = data
	return nil
}

func (s *MemoryStore) Get(ctx context.Context, key string, value any) error {
	data, ok := s.Data[dbPrefix+key]
	if !ok {
		return ErrKeyNotFound
	}

	return json.Unmarshal(data, value)
}
