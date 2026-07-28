package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type recordingPublisher struct {
	events []string
}

func (p *recordingPublisher) Publish(_ context.Context, channel string, _ fmt.Stringer) error {
	p.events = append(p.events, channel)
	return nil
}

func ptr[T any](v T) *T { return &v }

func TestAddItemValidation(t *testing.T) {
	t.Parallel()

	svc := New(nil, nil) // no DB needed: validation happens before any query
	_, err := svc.AddItem(context.Background(), nil, "", "1")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestAddItemByNameCreatesAndPublishes(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, category_id, name FROM item WHERE name =").
		WithArgs("Apples").
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name"}).
			AddRow(7, 2, "Apples"))
	mock.ExpectQuery("INSERT INTO item_list").
		WithArgs(int32(7), "3").
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "quantity", "done"}).
			AddRow(42, 7, "3", false))
	mock.ExpectQuery("SELECT id, category_id, name FROM item WHERE id =").
		WithArgs(int32(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name"}).
			AddRow(7, 2, "Apples"))
	mock.ExpectCommit()

	entry, err := svc.AddItem(context.Background(), nil, "Apples", "3")
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}

	if entry.ID != 42 || entry.ItemID != 7 || entry.Name != "Apples" || entry.Quantity != "3" || entry.Done {
		t.Fatalf("unexpected entry: %#v", entry)
	}
	if len(pub.events) != 1 || pub.events[0] != EventList {
		t.Fatalf("expected [%s] events, got %v", EventList, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAddItemUnknownIDIsNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, category_id, name FROM item WHERE id =").
		WithArgs(int32(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name"}))
	mock.ExpectRollback()

	_, err = svc.AddItem(context.Background(), ptr(int32(99)), "", "1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateItemNotOnListIsNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	// SummarizeItem returns a row with a NULL list_id (item not on the list).
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id FROM item").
		WithArgs(int32(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name", "category_name", "list_id"}).
			AddRow(5, 2, "Milk", "Dairy", nil))
	mock.ExpectRollback()

	_, err = svc.UpdateItem(context.Background(), 5, ptr("2"), nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateItemPartialInSingleTransaction(t *testing.T) {
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
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name", "category_name", "list_id"}).
			AddRow(5, 2, "Milk", "Dairy", 11))
	mock.ExpectQuery("SELECT item_list.id, item_list.item_id, item.name, item.category_id, item_list.quantity, item_list.done FROM item_list").
		WithArgs(int32(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "name", "category_id", "quantity", "done"}).
			AddRow(11, 5, "Milk", 2, "1", false))
	mock.ExpectQuery("UPDATE item_list SET").
		WithArgs(int32(5), "2", true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "quantity", "done"}).
			AddRow(11, 5, "2", true))
	mock.ExpectCommit()

	entry, err := svc.UpdateItem(context.Background(), 5, ptr("2"), ptr(true))
	if err != nil {
		t.Fatalf("UpdateItem: %v", err)
	}
	if entry.Quantity != "2" || !entry.Done || entry.Name != "Milk" {
		t.Fatalf("unexpected entry: %#v", entry)
	}
	// done changed -> list + cart events
	if len(pub.events) != 2 || pub.events[0] != EventList || pub.events[1] != EventCart {
		t.Fatalf("expected [%s %s] events, got %v", EventList, EventCart, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMarkDoneNotOnListIsNotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	// MarkItemDone matches zero rows -> sql.ErrNoRows from the scan.
	mock.ExpectQuery("UPDATE item_list SET done =").
		WithArgs(int32(5), true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "quantity", "done"}))

	err = svc.MarkDone(context.Background(), 5, true)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRemoveItemPublishesListEvent(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectExec("DELETE FROM item_list WHERE item_id =").
		WithArgs(int32(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.RemoveItem(context.Background(), 5); err != nil {
		t.Fatalf("RemoveItem: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0] != EventList {
		t.Fatalf("expected [%s] events, got %v", EventList, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFinishPublishesCartEvent(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := &recordingPublisher{}
	svc := New(db, pub)

	mock.ExpectExec("DELETE FROM item_list WHERE done = TRUE").
		WillReturnResult(sqlmock.NewResult(0, 3))

	if err := svc.Finish(context.Background()); err != nil {
		t.Fatalf("Finish: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0] != EventCart {
		t.Fatalf("expected [%s] events, got %v", EventCart, pub.events)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestNilPublisherIsNoOp(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectExec("DELETE FROM item_list WHERE done = TRUE").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.Finish(context.Background()); err != nil {
		t.Fatalf("Finish with nil publisher: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
