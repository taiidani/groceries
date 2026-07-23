package models

import (
	"context"
	"errors"
)

func (q *Queries) ValidateStore(ctx context.Context, s Store) error {
	var vErr error

	if len(s.Name) < 3 {
		vErr = errors.Join(vErr, errors.New("provided name needs to be at least 3 characters"))
	}

	// Check for existing Store
	if s.ID == 0 {
		_, err := q.GetStoreByName(ctx, s.Name)
		if err == nil {
			vErr = errors.Join(vErr, errors.New("store already found"))
		}
	}

	return vErr
}
