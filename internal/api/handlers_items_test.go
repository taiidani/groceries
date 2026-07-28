package api

import (
	"encoding/json"
	"testing"
)

func TestItemJSONMarshalIncludesCategoryName(t *testing.T) {
	t.Parallel()

	item := itemJSON{
		ID:           1,
		CategoryID:   2,
		CategoryName: "Produce",
		Name:         "Apples",
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
	if _, ok := payload["list"]; !ok {
		t.Fatalf("expected list key to be present, got %#v", payload)
	}
	if payload["list"] != nil {
		t.Fatalf("expected list to be null when unset, got %#v", payload["list"])
	}
}

func TestItemJSONMarshalIncludesList(t *testing.T) {
	t.Parallel()

	item := itemJSON{
		ID:           1,
		CategoryID:   2,
		CategoryName: "Produce",
		Name:         "Apples",
		List: &itemListJSON{
			ID:         7,
			ItemID:     1,
			CategoryID: "2",
			Quantity:   "3",
			Done:       false,
			Name:       "Apples",
		},
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal item: %v", err)
	}

	var payload struct {
		List struct {
			ID       int    `json:"id"`
			Quantity string `json:"quantity"`
			Done     bool   `json:"done"`
		} `json:"list"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload.List.ID != 7 || payload.List.Quantity != "3" || payload.List.Done {
		t.Fatalf("unexpected list payload: %#v", payload.List)
	}
}
