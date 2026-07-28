package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/taiidani/groceries/internal/db/models"
)

// CategoryDetail is a category together with the items that belong to it.
type CategoryDetail struct {
	models.Category
	Items []models.Item
}

// ListCategories returns all categories ordered by name.
func (s *Service) ListCategories(ctx context.Context) ([]models.Category, error) {
	categories, err := s.queries.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
}

// GetCategory returns a category and its items. Returns ErrNotFound if the
// category does not exist.
func (s *Service) GetCategory(ctx context.Context, id int32) (CategoryDetail, error) {
	category, err := s.queries.GetCategory(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return CategoryDetail{}, fmt.Errorf("category %d: %w", id, ErrNotFound)
	} else if err != nil {
		return CategoryDetail{}, fmt.Errorf("get category %d: %w", id, err)
	}

	items, err := s.queries.ListItemsForCategory(ctx, category.ID)
	if err != nil {
		return CategoryDetail{}, fmt.Errorf("list items for category %d: %w", id, err)
	}

	return CategoryDetail{Category: category, Items: items}, nil
}

// CreateCategory creates a new category in the given store. name and storeID
// are required (ErrValidation otherwise).
func (s *Service) CreateCategory(ctx context.Context, storeID int32, name, description string) (models.Category, error) {
	if name == "" {
		return models.Category{}, fmt.Errorf("name is required: %w", ErrValidation)
	}
	if storeID == 0 {
		return models.Category{}, fmt.Errorf("store_id is required: %w", ErrValidation)
	}

	category, err := s.queries.CreateCategory(ctx, models.CreateCategoryParams{
		StoreID:     storeID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return models.Category{}, fmt.Errorf("store %d: %w", storeID, ErrNotFound)
		}
		return models.Category{}, fmt.Errorf("create category: %w", err)
	}

	s.publish(ctx, EventCategory)
	return category, nil
}

// UpdateCategory updates an existing category. name and storeID are required
// (ErrValidation otherwise). Returns ErrNotFound if the category does not
// exist.
func (s *Service) UpdateCategory(ctx context.Context, id, storeID int32, name, description string) (models.Category, error) {
	if name == "" {
		return models.Category{}, fmt.Errorf("name is required: %w", ErrValidation)
	}
	if storeID == 0 {
		return models.Category{}, fmt.Errorf("store_id is required: %w", ErrValidation)
	}

	if _, err := s.queries.GetCategory(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return models.Category{}, fmt.Errorf("category %d: %w", id, ErrNotFound)
	} else if err != nil {
		return models.Category{}, fmt.Errorf("get category %d: %w", id, err)
	}

	category, err := s.queries.UpdateCategory(ctx, models.UpdateCategoryParams{
		ID:          id,
		StoreID:     storeID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return models.Category{}, fmt.Errorf("store %d: %w", storeID, ErrNotFound)
		}
		return models.Category{}, fmt.Errorf("update category %d: %w", id, err)
	}

	s.publish(ctx, EventCategory)
	return category, nil
}

// DeleteCategory removes a category. Returns ErrNotFound if the category does
// not exist, or ErrConflict if it is still referenced by items.
func (s *Service) DeleteCategory(ctx context.Context, id int32) error {
	if _, err := s.queries.GetCategory(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("category %d: %w", id, ErrNotFound)
	} else if err != nil {
		return fmt.Errorf("get category %d: %w", id, err)
	}

	if err := s.queries.DeleteCategory(ctx, id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return fmt.Errorf("category %d is still in use: %w", id, ErrConflict)
		}
		return fmt.Errorf("delete category %d: %w", id, err)
	}

	s.publish(ctx, EventCategory)
	return nil
}
