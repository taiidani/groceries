package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/taiidani/groceries/internal/db/models"
)

// pgForeignKeyViolation is the PostgreSQL error code for foreign key
// constraint violations.
const pgForeignKeyViolation = "23503"

// Item is a catalog item along with its category name and shopping list
// state. The List fields are only meaningful when OnList is true.
type Item struct {
	ID           int32
	CategoryID   int32
	CategoryName string
	Name         string

	// OnList reports whether the item is currently on the shopping list.
	OnList bool
	// ListID is the item_list entry ID, valid only when OnList is true.
	ListID int32
	// ListQuantity is the quantity recorded on the list, valid only when
	// OnList is true.
	ListQuantity string
	// ListDone reports whether the list entry is marked done, valid only
	// when OnList is true.
	ListDone bool
}

// ItemFilters narrows the result of ListItems. A nil field leaves the
// corresponding dimension unfiltered.
type ItemFilters struct {
	CategoryID *int32
	InList     *bool
}

// ListItems returns catalog items with category and list state, optionally
// filtered by category and/or list membership.
func (s *Service) ListItems(ctx context.Context, filters ItemFilters) ([]Item, error) {
	rows, err := s.queries.SummarizeItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("summarize items: %w", err)
	}

	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		if filters.CategoryID != nil && row.CategoryID != *filters.CategoryID {
			continue
		}
		if filters.InList != nil && row.ListID.Valid != *filters.InList {
			continue
		}

		items = append(items, itemFromSummaryRow(row))
	}

	return items, nil
}

// GetItem returns a single catalog item with category and list state.
// Returns ErrNotFound if the item does not exist.
func (s *Service) GetItem(ctx context.Context, id int32) (Item, error) {
	row, err := s.queries.SummarizeItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, fmt.Errorf("item %d: %w", id, ErrNotFound)
	} else if err != nil {
		return Item{}, fmt.Errorf("summarize item %d: %w", id, err)
	}

	item := Item{
		ID:           row.ID,
		CategoryID:   row.CategoryID,
		CategoryName: row.CategoryName.String,
		Name:         row.Name,
	}

	if row.ListID.Valid {
		entry, err := s.queries.GetListItem(ctx, row.ListID.Int32)
		if err != nil {
			return Item{}, fmt.Errorf("get list item %d: %w", row.ListID.Int32, err)
		}
		item.OnList = true
		item.ListID = entry.ID
		item.ListQuantity = entry.Quantity
		item.ListDone = entry.Done
	}

	return item, nil
}

// CreateItem creates a new catalog item. name and categoryID are required;
// returns ErrValidation otherwise. Publishes EventList on success.
func (s *Service) CreateItem(ctx context.Context, categoryID int32, name string) (Item, error) {
	if name == "" {
		return Item{}, fmt.Errorf("name is required: %w", ErrValidation)
	}
	if categoryID == 0 {
		return Item{}, fmt.Errorf("category_id is required: %w", ErrValidation)
	}

	var item Item
	err := s.tx(ctx, func(q *models.Queries) error {
		created, err := q.CreateItem(ctx, models.CreateItemParams{
			CategoryID: categoryID,
			Name:       name,
		})
		if err != nil {
			return fmt.Errorf("create item: %w", err)
		}

		// Re-fetch the summary so the response includes category_name
		row, err := q.SummarizeItem(ctx, created.ID)
		if err != nil {
			return fmt.Errorf("summarize item %d: %w", created.ID, err)
		}

		item = Item{
			ID:           row.ID,
			CategoryID:   row.CategoryID,
			CategoryName: row.CategoryName.String,
			Name:         row.Name,
		}
		return nil
	})
	if err != nil {
		return Item{}, err
	}

	s.publish(ctx, EventList)
	return item, nil
}

// UpdateCatalogItem updates a catalog item's name and category. When
// listQuantity is non-nil and the item is on the shopping list, the list
// entry's quantity is updated in the same transaction (absorbing what used
// to be two separate API calls from the web layer). Returns ErrNotFound if
// the item does not exist. Publishes EventList on success.
func (s *Service) UpdateCatalogItem(ctx context.Context, id int32, name string, categoryID int32, listQuantity *string) (Item, error) {
	if name == "" {
		return Item{}, fmt.Errorf("name is required: %w", ErrValidation)
	}
	if categoryID == 0 {
		return Item{}, fmt.Errorf("category_id is required: %w", ErrValidation)
	}

	var item Item
	err := s.tx(ctx, func(q *models.Queries) error {
		summary, err := q.SummarizeItem(ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("item %d: %w", id, ErrNotFound)
		} else if err != nil {
			return fmt.Errorf("summarize item %d: %w", id, err)
		}

		if _, err := q.UpdateItem(ctx, models.UpdateItemParams{
			ID:         id,
			CategoryID: categoryID,
			Name:       name,
		}); err != nil {
			return fmt.Errorf("update item %d: %w", id, err)
		}

		item = Item{
			ID:         id,
			CategoryID: categoryID,
			Name:       name,
			OnList:     summary.ListID.Valid,
		}

		if summary.ListID.Valid {
			entry, err := q.GetListItem(ctx, summary.ListID.Int32)
			if err != nil {
				return fmt.Errorf("get list item %d: %w", summary.ListID.Int32, err)
			}

			quantity := entry.Quantity
			done := entry.Done
			if listQuantity != nil {
				quantity = *listQuantity
			}
			if quantity != entry.Quantity {
				updated, err := q.UpdateListItem(ctx, models.UpdateListItemParams{
					ItemID:   id,
					Quantity: quantity,
					Done:     entry.Done,
				})
				if err != nil {
					return fmt.Errorf("update list item: %w", err)
				}
				quantity = updated.Quantity
				done = updated.Done
			}

			item.ListID = entry.ID
			item.ListQuantity = quantity
			item.ListDone = done
		}

		return nil
	})
	if err != nil {
		return Item{}, err
	}

	// CategoryName is not part of the update payload; resolve it from the
	// (already committed) state for the convenience of callers.
	if resolved, err := s.queries.SummarizeItem(ctx, id); err == nil {
		item.CategoryName = resolved.CategoryName.String
	}

	s.publish(ctx, EventList)
	return item, nil
}

// DeleteItem removes a catalog item. Returns ErrNotFound if the item does
// not exist, and ErrConflict if it is still referenced by other records
// (foreign key violation). Publishes EventList on success.
func (s *Service) DeleteItem(ctx context.Context, id int32) error {
	if _, err := s.queries.SummarizeItem(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("item %d: %w", id, ErrNotFound)
		}
		return fmt.Errorf("summarize item %d: %w", id, err)
	}

	if err := s.queries.DeleteItem(ctx, id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return fmt.Errorf("item %d is in use: %w", id, ErrConflict)
		}
		return fmt.Errorf("delete item %d: %w", id, err)
	}

	s.publish(ctx, EventList)
	return nil
}

func itemFromSummaryRow(row models.SummarizeItemsRow) Item {
	item := Item{
		ID:           row.ID,
		CategoryID:   row.CategoryID,
		CategoryName: row.CategoryName.String,
		Name:         row.Name,
	}
	if row.ListID.Valid {
		item.OnList = true
		item.ListID = row.ListID.Int32
		item.ListQuantity = row.ListQuantity.String
		item.ListDone = row.ListDone.Bool
	}
	return item
}
