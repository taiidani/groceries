package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/taiidani/groceries/internal/db/models"
)

// pgUniqueViolation is the PostgreSQL error code for unique constraint
// violations.
const pgUniqueViolation = "23505"

// ListEntry is a single item on the shopping list.
type ListEntry struct {
	// ID is the list entry (item_list) ID.
	ID int32
	// ItemID is the ID of the underlying catalog item.
	ItemID     int32
	Name       string
	CategoryID int32
	Quantity   string
	Done       bool
}

// ListSummary is the shopping list plus aggregate counts.
type ListSummary struct {
	Items     []ListEntry
	Total     int
	TotalDone int
}

// GetList returns the shopping list with aggregate totals.
func (s *Service) GetList(ctx context.Context) (ListSummary, error) {
	rows, err := s.queries.LoadList(ctx)
	if err != nil {
		return ListSummary{}, fmt.Errorf("load list: %w", err)
	}

	summary := ListSummary{Items: make([]ListEntry, 0, len(rows))}
	for _, row := range rows {
		if row.ListDone {
			summary.TotalDone++
		}
		summary.Items = append(summary.Items, ListEntry{
			ID:         row.ListID,
			ItemID:     row.ID,
			Name:       row.Name,
			CategoryID: row.CategoryID,
			Quantity:   row.ListQuantity,
			Done:       row.ListDone,
		})
	}
	summary.Total = len(rows)

	return summary, nil
}

// AddItem adds an item to the shopping list. Exactly one of itemID or name
// must be provided:
//
//   - itemID set: the catalog item must already exist (ErrNotFound otherwise).
//   - name set: the catalog item is looked up by name and created on the fly
//     if missing.
//
// Returns ErrConflict if the item is already on the list.
func (s *Service) AddItem(ctx context.Context, itemID *int32, name string, quantity string) (ListEntry, error) {
	if itemID == nil && name == "" {
		return ListEntry{}, fmt.Errorf("one of itemID or name is required: %w", ErrValidation)
	}

	var entry ListEntry
	err := s.tx(ctx, func(q *models.Queries) error {
		var id int32
		switch {
		case itemID != nil:
			item, err := q.GetItem(ctx, *itemID)
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("item %d: %w", *itemID, ErrNotFound)
			} else if err != nil {
				return fmt.Errorf("get item: %w", err)
			}
			id = item.ID

		default: // name != ""
			item, err := q.GetItemByName(ctx, name)
			if errors.Is(err, sql.ErrNoRows) {
				// The item doesn't exist yet. Create it on the fly.
				item, err = q.CreateItem(ctx, models.CreateItemParams{Name: name})
			}
			if err != nil {
				return fmt.Errorf("get or create item %q: %w", name, err)
			}
			id = item.ID
		}

		row, err := q.CreateListItem(ctx, models.CreateListItemParams{
			ItemID:   id,
			Quantity: quantity,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
				return fmt.Errorf("item %d is already on the list: %w", id, ErrConflict)
			}
			return fmt.Errorf("create list item: %w", err)
		}

		resolved, err := q.GetItem(ctx, id)
		if err != nil {
			return fmt.Errorf("resolve item %d: %w", id, err)
		}

		entry = ListEntry{
			ID:         row.ID,
			ItemID:     row.ItemID,
			Name:       resolved.Name,
			CategoryID: resolved.CategoryID,
			Quantity:   row.Quantity,
			Done:       row.Done,
		}
		return nil
	})
	if err != nil {
		return ListEntry{}, err
	}

	s.publish(ctx, EventList)
	return entry, nil
}

// UpdateItem applies a partial update to a list entry identified by item ID.
// quantity and done are optional; only non-nil fields are changed. Returns
// ErrNotFound if the item does not exist or is not on the list.
func (s *Service) UpdateItem(ctx context.Context, itemID int32, quantity *string, done *bool) (ListEntry, error) {
	var entry ListEntry
	err := s.tx(ctx, func(q *models.Queries) error {
		summary, err := q.SummarizeItem(ctx, itemID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("item %d: %w", itemID, ErrNotFound)
		} else if err != nil {
			return fmt.Errorf("summarize item: %w", err)
		}

		if !summary.ListID.Valid {
			return fmt.Errorf("item %d is not on the list: %w", itemID, ErrNotFound)
		}

		listItem, err := q.GetListItem(ctx, summary.ListID.Int32)
		if err != nil {
			return fmt.Errorf("get list item %d: %w", summary.ListID.Int32, err)
		}

		newQuantity := listItem.Quantity
		if quantity != nil {
			newQuantity = *quantity
		}
		newDone := listItem.Done
		if done != nil {
			newDone = *done
		}

		row, err := q.UpdateListItem(ctx, models.UpdateListItemParams{
			ItemID:   itemID,
			Quantity: newQuantity,
			Done:     newDone,
		})
		if err != nil {
			return fmt.Errorf("update list item: %w", err)
		}

		entry = ListEntry{
			ID:         row.ID,
			ItemID:     row.ItemID,
			Name:       summary.Name,
			CategoryID: summary.CategoryID,
			Quantity:   row.Quantity,
			Done:       row.Done,
		}
		return nil
	})
	if err != nil {
		return ListEntry{}, err
	}

	events := []string{EventList}
	if done != nil {
		events = append(events, EventCart)
	}
	s.publish(ctx, events...)
	return entry, nil
}

// RemoveItem removes an item from the shopping list by item ID.
func (s *Service) RemoveItem(ctx context.Context, itemID int32) error {
	if err := s.queries.DeleteListItemByItemID(ctx, itemID); err != nil {
		return fmt.Errorf("delete list item %d: %w", itemID, err)
	}

	s.publish(ctx, EventList)
	return nil
}

// MarkDone marks an item on the list as done or not done. Returns
// ErrNotFound if the item is not on the list.
func (s *Service) MarkDone(ctx context.Context, itemID int32, done bool) error {
	_, err := s.queries.MarkItemDone(ctx, models.MarkItemDoneParams{ItemID: itemID, Done: done})
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("item %d is not on the list: %w", itemID, ErrNotFound)
	} else if err != nil {
		return fmt.Errorf("mark item %d done: %w", itemID, err)
	}

	s.publish(ctx, EventList, EventCart)
	return nil
}

// Finish removes all done items from the shopping list.
func (s *Service) Finish(ctx context.Context) error {
	if err := s.queries.FinishShopping(ctx); err != nil {
		return fmt.Errorf("finish shopping: %w", err)
	}

	s.publish(ctx, EventCart)
	return nil
}
