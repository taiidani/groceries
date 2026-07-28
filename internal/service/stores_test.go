package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func storeColumns() []string {
	return []string{"id", "name"}
}

func TestCreateStoreValidation(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	// ValidateStore still performs the name lookup even when the length
	// check fails.
	mock.ExpectQuery("SELECT id, name FROM store WHERE name =").
		WithArgs("ab").
		WillReturnRows(sqlmock.NewRows(storeColumns()))

	if _, err := svc.CreateStore(context.Background(), "ab"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for short name, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateStoreDuplicateIsValidation(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	// ValidateStore looks the name up and finds an existing store.
	mock.ExpectQuery("SELECT id, name FROM store WHERE name =").
		WithArgs("Costco").
		WillReturnRows(sqlmock.NewRows(storeColumns()).AddRow(1, "Costco"))

	_, err = svc.CreateStore(context.Background(), "Costco")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateStorePublishesNothing(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectQuery("SELECT id, name FROM store WHERE name =").
		WithArgs("Costco").
		WillReturnRows(sqlmock.NewRows(storeColumns()))
	mock.ExpectQuery("INSERT INTO store").
		WithArgs("Costco").
		WillReturnRows(sqlmock.NewRows(storeColumns()).AddRow(3, "Costco"))

	store, err := svc.CreateStore(context.Background(), "Costco")
	if err != nil {
		t.Fatalf("CreateStore: %v", err)
	}
	if store.ID != 3 || store.Name != "Costco" {
		t.Fatalf("unexpected store: %#v", store)
	}
	if len(pub.events) != 0 {
		t.Fatalf("expected no store events, got %v", pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetStoreNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name FROM store WHERE id =").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(storeColumns()))

	_, err = svc.GetStore(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetStoreIncludesCategories(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name FROM store WHERE id =").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows(storeColumns()).AddRow(1, "Costco"))
	mock.ExpectQuery("SELECT id, name, description, store_id FROM category WHERE store_id =").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows(categoryColumns()).
			AddRow(5, "Produce", "Fresh", 1))

	detail, err := svc.GetStore(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetStore: %v", err)
	}
	if detail.Name != "Costco" || len(detail.Categories) != 1 || detail.Categories[0].Name != "Produce" {
		t.Fatalf("unexpected detail: %#v", detail)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateStoreNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name FROM store WHERE id =").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(storeColumns()))

	_, err = svc.UpdateStore(context.Background(), 99, "Costco")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteStoreInUseIsConflict(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name FROM store WHERE id =").
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows(storeColumns()).AddRow(1, "Costco"))
	mock.ExpectExec("DELETE FROM store WHERE id =").
		WithArgs(int32(1)).
		WillReturnError(&pgconn.PgError{Code: pgForeignKeyViolation, Message: "violates foreign key constraint"})

	err = svc.DeleteStore(context.Background(), 1)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteStoreNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name FROM store WHERE id =").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(storeColumns()))

	err = svc.DeleteStore(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
