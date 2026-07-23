package models

import (
	"context"
	"errors"
)

func (q *Queries) ValidateGroup(ctx context.Context, g Group) error {
	var vErr error

	if len(g.Name) < 3 {
		vErr = errors.Join(vErr, errors.New("provided name needs to be at least 3 characters"))
	}

	return vErr
}
