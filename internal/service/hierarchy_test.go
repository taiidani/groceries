package service

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func expectHierarchyQueries(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT id, name FROM store").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Corner Market").
			AddRow(2, "Mega Mart"))
	mock.ExpectQuery("SELECT id, name, description, store_id FROM category").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "store_id"}).
			AddRow(10, "Produce", "Fresh things", 1).
			AddRow(11, "Dairy", "Moo", 1).
			AddRow(12, "Bakery", "Bread", 2))
}

func hierarchyItemRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "category_id", "category_name", "list_id", "list_quantity", "list_done"}).
		AddRow(100, "Apples", 10, "Produce", 5, "3", false).
		AddRow(101, "Milk", 11, "Dairy", nil, nil, nil).
		AddRow(102, "Baguette", 12, "Bakery", 6, "1", true)
}

func TestLoadStoreHierarchyFull(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	expectHierarchyQueries(mock)
	mock.ExpectQuery("SELECT item.id, item.name, item.category_id, category.name AS category_name").
		WillReturnRows(hierarchyItemRows())

	stores, err := svc.LoadStoreHierarchy(context.Background(), HierarchyInput{})
	if err != nil {
		t.Fatalf("LoadStoreHierarchy: %v", err)
	}

	if len(stores) != 2 {
		t.Fatalf("expected 2 stores, got %d", len(stores))
	}

	if stores[0].Name != "Corner Market" || len(stores[0].Categories) != 2 {
		t.Fatalf("unexpected first store: %#v", stores[0])
	}
	produce := stores[0].Categories[0]
	if produce.Name != "Produce" || len(produce.Items) != 1 {
		t.Fatalf("unexpected produce category: %#v", produce)
	}
	apples := produce.Items[0]
	if apples.Name != "Apples" || apples.List == nil || apples.List.Quantity != "3" || apples.List.Done {
		t.Fatalf("unexpected apples item: %#v", apples)
	}

	dairy := stores[0].Categories[1]
	if len(dairy.Items) != 1 || dairy.Items[0].List != nil {
		t.Fatalf("expected unlisted milk, got %#v", dairy.Items)
	}

	if len(stores[1].Categories) != 1 || len(stores[1].Categories[0].Items) != 1 {
		t.Fatalf("unexpected second store: %#v", stores[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLoadStoreHierarchyListOnlyExcludingDoneAndEmpty(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	expectHierarchyQueries(mock)
	mock.ExpectQuery("SELECT item.id, item.name, item.category_id, category.name AS category_name").
		WillReturnRows(hierarchyItemRows())

	stores, err := svc.LoadStoreHierarchy(context.Background(), HierarchyInput{
		OnlyListItems:         true,
		ExcludeDoneItems:      true,
		ExcludeEmptyGroupings: true,
	})
	if err != nil {
		t.Fatalf("LoadStoreHierarchy: %v", err)
	}

	// Only "Corner Market" -> "Produce" -> "Apples" survives: milk is not on
	// the list and the baguette is done, which empties "Mega Mart".
	if len(stores) != 1 {
		t.Fatalf("expected 1 store, got %d", len(stores))
	}
	if len(stores[0].Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(stores[0].Categories))
	}
	cat := stores[0].Categories[0]
	if cat.Name != "Produce" || len(cat.Items) != 1 || cat.Items[0].Name != "Apples" {
		t.Fatalf("unexpected category contents: %#v", cat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLoadStoreHierarchyEmptyIsNotNil(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := New(db, nil)

	mock.ExpectQuery("SELECT id, name FROM store").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	mock.ExpectQuery("SELECT id, name, description, store_id FROM category").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "store_id"}))
	mock.ExpectQuery("SELECT item.id, item.name, item.category_id, category.name AS category_name").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "category_id", "category_name", "list_id", "list_quantity", "list_done"}))

	stores, err := svc.LoadStoreHierarchy(context.Background(), HierarchyInput{ExcludeEmptyGroupings: true})
	if err != nil {
		t.Fatalf("LoadStoreHierarchy: %v", err)
	}
	if stores == nil || len(stores) != 0 {
		t.Fatalf("expected empty non-nil slice, got %#v", stores)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
