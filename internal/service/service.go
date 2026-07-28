// Package service contains the shared domain logic for the groceries
// application. Both transports (internal/api JSON handlers and internal/server
// HTML handlers) delegate business logic — validation, DB call sequencing,
// error semantics, and event publishing — to this package.
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/taiidani/groceries/internal/db/models"
)

// Typed sentinel errors. Domain errors wrap these with %w, so transports
// should classify failures with errors.Is:
//
//	ErrValidation -> 400 Bad Request
//	ErrNotFound   -> 404 Not Found
//	ErrConflict   -> 409 Conflict
var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation error")
)

// SSE event names published by the service layer. These are the single source
// of truth; transports should reference them instead of defining their own.
const (
	// EventList is triggered when an item is added to or removed from the list
	EventList = "list"

	// EventCart is triggered when an item is added to or removed from the cart
	EventCart = "cart"

	// EventCategory is triggered when a category is added or removed
	EventCategory = "category"
)

// Publisher is the minimal event-publishing surface the service needs.
// events.PubSub satisfies this interface directly.
type Publisher interface {
	Publish(ctx context.Context, channel string, data fmt.Stringer) error
}

// Service holds the shared dependencies used by all domain services.
type Service struct {
	db      *sql.DB
	queries *models.Queries
	pub     Publisher
}

// New creates a Service backed by the given database handle. pub may be nil,
// in which case event publishing is a no-op.
func New(conn *sql.DB, pub Publisher) *Service {
	return &Service{
		db:      conn,
		queries: models.New(conn),
		pub:     pub,
	}
}

// tx runs fn inside a database transaction, committing on success and rolling
// back on error.
func (s *Service) tx(ctx context.Context, fn func(q *models.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// publish fires an SSE event. Publishing is best-effort: failures are logged
// but do not fail the underlying operation. It is a no-op when no Publisher
// was configured.
func (s *Service) publish(ctx context.Context, events ...string) {
	if s.pub == nil {
		return
	}
	for _, event := range events {
		if err := s.pub.Publish(ctx, event, nil); err != nil {
			slog.WarnContext(ctx, "failed to publish event", "event", event, "error", err)
		}
	}
}
