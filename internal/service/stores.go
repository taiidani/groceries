package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/taiidani/groceries/internal/db/models"
)

// StoreDetail is a store together with the categories that belong to it.
type StoreDetail struct {
	models.Store
	Categories []models.Category
}

// ListStores returns all stores ordered by name.
func (s *Service) ListStores(ctx context.Context) ([]models.Store, error) {
	stores, err := s.queries.ListStores(ctx)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	return stores, nil
}

// GetStore returns a store and its categories. Returns ErrNotFound if the
// store does not exist.
func (s *Service) GetStore(ctx context.Context, id int32) (StoreDetail, error) {
	store, err := s.queries.GetStore(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return StoreDetail{}, fmt.Errorf("store %d: %w", id, ErrNotFound)
	} else if err != nil {
		return StoreDetail{}, fmt.Errorf("get store %d: %w", id, err)
	}

	categories, err := s.queries.ListCategoriesForStore(ctx, store.ID)
	if err != nil {
		return StoreDetail{}, fmt.Errorf("list categories for store %d: %w", id, err)
	}

	return StoreDetail{Store: store, Categories: categories}, nil
}

// CreateStore creates a new store. The name must be at least 3 characters
// and unique (ErrValidation otherwise). Stores do not publish events.
func (s *Service) CreateStore(ctx context.Context, name string) (models.Store, error) {
	if err := s.queries.ValidateStore(ctx, models.Store{Name: name}); err != nil {
		return models.Store{}, fmt.Errorf("%w: %w", ErrValidation, err)
	}

	store, err := s.queries.CreateStore(ctx, name)
	if err != nil {
		return models.Store{}, fmt.Errorf("create store: %w", err)
	}

	return store, nil
}

// UpdateStore renames an existing store. The name must be at least 3
// characters (ErrValidation otherwise). Returns ErrNotFound if the store
// does not exist. Stores do not publish events.
func (s *Service) UpdateStore(ctx context.Context, id int32, name string) (models.Store, error) {
	if _, err := s.queries.GetStore(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return models.Store{}, fmt.Errorf("store %d: %w", id, ErrNotFound)
	} else if err != nil {
		return models.Store{}, fmt.Errorf("get store %d: %w", id, err)
	}

	if err := s.queries.ValidateStore(ctx, models.Store{ID: id, Name: name}); err != nil {
		return models.Store{}, fmt.Errorf("%w: %w", ErrValidation, err)
	}

	store, err := s.queries.UpdateStore(ctx, models.UpdateStoreParams{ID: id, Name: name})
	if err != nil {
		return models.Store{}, fmt.Errorf("update store %d: %w", id, err)
	}

	return store, nil
}

// DeleteStore removes a store. Returns ErrNotFound if the store does not
// exist, or ErrConflict if it is still referenced by categories.
func (s *Service) DeleteStore(ctx context.Context, id int32) error {
	if _, err := s.queries.GetStore(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("store %d: %w", id, ErrNotFound)
	} else if err != nil {
		return fmt.Errorf("get store %d: %w", id, err)
	}

	if err := s.queries.DeleteStore(ctx, id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return fmt.Errorf("store %d is still in use: %w", id, ErrConflict)
		}
		return fmt.Errorf("delete store %d: %w", id, err)
	}

	return nil
}
