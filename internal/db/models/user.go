package models

import (
	"context"
	"errors"
)

func (q *Queries) ValidateUser(ctx context.Context, u User) error {
	var vErr error
	if u.Name == "" {
		vErr = errors.Join(vErr, errors.New("username must be valid"))
	}

	return vErr
}
