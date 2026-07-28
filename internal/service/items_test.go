package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func summarizeItemsColumns() []string {
	return []string{"id", "name", "category_id", "category_name", "list_id", "list_quantity", "list_done"}
}

func summarizeItemColumns() []string {
	return []string{"id", "category_id", "name", "category_name", "list_id"}
}

func TestListItemsFilters(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT item.id, item.name, item.category_id, category.name AS category_name").
		WillReturnRows(sqlmock.NewRows(summarizeItemsColumns()).
			AddRow(1, "Apples", 2, "Produce", 10, "3", false).
			AddRow(2, "Milk", 3, "Dairy", nil, nil, nil).
			AddRow(3, "Bread", 3, "Bakery", 11, "1", true))

	categoryID := int32(3)
	inList := true
	items, err := svc.ListItems(context.Background(), ItemFilters{CategoryID: &categoryID, InList: &inList})
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.ID != 3 || item.Name != "Bread" || !item.OnList || item.ListID != 11 ||
		item.ListQuantity != "1" || !item.ListDone {
		t.Fatalf("unexpected item: %#v", item)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetItemNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()))

	_, err = svc.GetItem(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetItemWithList(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()).
			AddRow(5, 2, "Milk", "Dairy", 11))
	mock.ExpectQuery("SELECT item_list.id, item_list.item_id, item.name, item.category_id, item_list.quantity, item_list.done FROM item_list").
		WithArgs(int32(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "name", "category_id", "quantity", "done"}).
			AddRow(11, 5, "Milk", 2, "2", false))

	item, err := svc.GetItem(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if !item.OnList || item.ListID != 11 || item.ListQuantity != "2" || item.CategoryName != "Dairy" {
		t.Fatalf("unexpected item: %#v", item)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateItemValidation(t *testing.T) {
	t.Parallel()

	svc := New(nil, nil) // no DB needed: validation happens before any query

	if _, err := svc.CreateItem(context.Background(), 0, "Apples"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for missing category, got %v", err)
	}
	if _, err := svc.CreateItem(context.Background(), 2, ""); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for missing name, got %v", err)
	}
}

func TestCreateItemPublishesListEvent(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO item").
		WithArgs(int32(2), "Apples").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name"}).
			AddRow(7, 2, "Apples"))
	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(7)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()).
			AddRow(7, 2, "Apples", "Produce", nil))
	mock.ExpectCommit()

	item, err := svc.CreateItem(context.Background(), 2, "Apples")
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}
	if item.ID != 7 || item.CategoryName != "Produce" || item.OnList {
		t.Fatalf("unexpected item: %#v", item)
	}
	if len(pub.events) != 1 || pub.events[0] != EventList {
		t.Fatalf("expected [%s] events, got %v", EventList, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateCatalogItemNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()))
	mock.ExpectRollback()

	_, err = svc.UpdateCatalogItem(context.Background(), 99, "Milk", 2, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateCatalogItemWithQuantityInSingleTransaction(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()).
			AddRow(5, 2, "Milk", "Dairy", 11))
	mock.ExpectQuery("UPDATE item SET").
		WithArgs(int32(5), int32(2), "Whole Milk").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name"}).
			AddRow(5, 2, "Whole Milk"))
	mock.ExpectQuery("SELECT item_list.id, item_list.item_id, item.name, item.category_id, item_list.quantity, item_list.done FROM item_list").
		WithArgs(int32(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "name", "category_id", "quantity", "done"}).
			AddRow(11, 5, "Milk", 2, "1", false))
	mock.ExpectQuery("UPDATE item_list SET").
		WithArgs(int32(5), "3", false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "quantity", "done"}).
			AddRow(11, 5, "3", false))
	mock.ExpectCommit()
	// Post-commit category name resolution
	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()).
			AddRow(5, 2, "Whole Milk", "Dairy", 11))

	item, err := svc.UpdateCatalogItem(context.Background(), 5, "Whole Milk", 2, ptr("3"))
	if err != nil {
		t.Fatalf("UpdateCatalogItem: %v", err)
	}
	if item.Name != "Whole Milk" || item.CategoryName != "Dairy" || !item.OnList || item.ListQuantity != "3" {
		t.Fatalf("unexpected item: %#v", item)
	}
	if len(pub.events) != 1 || pub.events[0] != EventList {
		t.Fatalf("expected [%s] events, got %v", EventList, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteItemNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()))

	err = svc.DeleteItem(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteItemForeignKeyViolationIsConflict(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()).
			AddRow(5, 2, "Milk", "Dairy", nil))
	mock.ExpectExec("DELETE FROM item WHERE id =").
		WithArgs(int32(5)).
		WillReturnError(&pgconn.PgError{Code: pgForeignKeyViolation})

	err = svc.DeleteItem(context.Background(), 5)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteItemOtherErrorIsInternal(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()).
			AddRow(5, 2, "Milk", "Dairy", nil))
	mock.ExpectExec("DELETE FROM item WHERE id =").
		WithArgs(int32(5)).
		WillReturnError(sql.ErrConnDone)

	err = svc.DeleteItem(context.Background(), 5)
	if err == nil || errors.Is(err, ErrConflict) || errors.Is(err, ErrNotFound) {
		t.Fatalf("expected unclassified error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteItemPublishesListEvent(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows(summarizeItemColumns()).
			AddRow(5, 2, "Milk", "Dairy", nil))
	mock.ExpectExec("DELETE FROM item WHERE id =").
		WithArgs(int32(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.DeleteItem(context.Background(), 5); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0] != EventList {
		t.Fatalf("expected [%s] events, got %v", EventList, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
