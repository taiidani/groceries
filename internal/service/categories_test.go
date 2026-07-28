package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func categoryColumns() []string {
	return []string{"id", "name", "description", "store_id"}
}

func TestCreateCategoryValidation(t *testing.T) {
	t.Parallel()

	svc := New(nil, nil) // validation happens before any query
	if _, err := svc.CreateCategory(context.Background(), 1, "", ""); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for empty name, got %v", err)
	}
	if _, err := svc.CreateCategory(context.Background(), 0, "Produce", ""); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for missing store, got %v", err)
	}
}

func TestCreateCategoryPublishesEvent(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectQuery("INSERT INTO category").
		WithArgs("Produce", int32(1), "Fresh").
		WillReturnRows(sqlmock.NewRows(categoryColumns()).
			AddRow(5, "Produce", "Fresh", 1))

	cat, err := svc.CreateCategory(context.Background(), 1, "Produce", "Fresh")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if cat.ID != 5 || cat.Name != "Produce" || cat.StoreID != 1 {
		t.Fatalf("unexpected category: %#v", cat)
	}
	if len(pub.events) != 1 || pub.events[0] != EventCategory {
		t.Fatalf("expected [%s] events, got %v", EventCategory, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetCategoryNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name, description, store_id FROM category WHERE id =").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()))

	_, err = svc.GetCategory(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetCategoryIncludesItems(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name, description, store_id FROM category WHERE id =").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()).
			AddRow(5, "Produce", "Fresh", 1))
	mock.ExpectQuery("SELECT id, category_id, name FROM item WHERE category_id =").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name"}).
			AddRow(7, 5, "Apples").
			AddRow(8, 5, "Bananas"))

	detail, err := svc.GetCategory(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetCategory: %v", err)
	}
	if detail.Name != "Produce" || len(detail.Items) != 2 || detail.Items[0].Name != "Apples" {
		t.Fatalf("unexpected detail: %#v", detail)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateCategoryNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name, description, store_id FROM category WHERE id =").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()))

	_, err = svc.UpdateCategory(context.Background(), 99, 1, "Produce", "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteCategoryInUseIsConflict(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectQuery("SELECT id, name, description, store_id FROM category WHERE id =").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()).
			AddRow(5, "Produce", "Fresh", 1))
	mock.ExpectExec("DELETE FROM category WHERE id =").
		WithArgs(int32(5)).
		WillReturnError(&pgconn.PgError{Code: pgForeignKeyViolation, Message: "violates foreign key constraint"})

	err = svc.DeleteCategory(context.Background(), 5)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if len(pub.events) != 0 {
		t.Fatalf("expected no events on failed delete, got %v", pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteCategoryPublishesEvent(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectQuery("SELECT id, name, description, store_id FROM category WHERE id =").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()).
			AddRow(5, "Produce", "Fresh", 1))
	mock.ExpectExec("DELETE FROM category WHERE id =").
		WithArgs(int32(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.DeleteCategory(context.Background(), 5); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0] != EventCategory {
		t.Fatalf("expected [%s] events, got %v", EventCategory, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteCategoryNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name, description, store_id FROM category WHERE id =").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()))

	err = svc.DeleteCategory(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
