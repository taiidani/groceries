package models

import (
	"encoding/json"
	"testing"
)

func TestItemMarshalJSONIncludesCategoryName(t *testing.T) {
	t.Parallel()

	item := Item{
		ID:           1,
		CategoryID:   2,
		Name:         "Apples",
		categoryName: "Produce",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal item: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload["category_name"] != "Produce" {
		t.Fatalf("expected category_name to be %q, got %#v", "Produce", payload["category_name"])
	}
}
