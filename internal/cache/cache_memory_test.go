package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore_SetAndGet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      interface{}
		expiration time.Duration
		wantErr    bool
	}{
		{
			name:       "string value",
			key:        "test-key",
			value:      "test-value",
			expiration: time.Minute,
			wantErr:    false,
		},
		{
			name:       "struct value",
			key:        "user-key",
			value:      struct{ Name string }{Name: "John"},
			expiration: time.Hour,
			wantErr:    false,
		},
		{
			name:       "int value",
			key:        "count",
			value:      42,
			expiration: time.Second,
			wantErr:    false,
		},
		{
			name:       "map value",
			key:        "data",
			value:      map[string]string{"foo": "bar"},
			expiration: time.Minute,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := &MemoryStore{Data: make(map[string][]byte)}

			// Set value
			err := store.Set(ctx, tt.key, tt.value, tt.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Get value
			var result interface{}
			switch tt.value.(type) {
			case string:
				var s string
				err = store.Get(ctx, tt.key, &s)
				result = s
			case int:
				var i int
				err = store.Get(ctx, tt.key, &i)
				result = i
			case map[string]string:
				var m map[string]string
				err = store.Get(ctx, tt.key, &m)
				result = m
			default:
				err = store.Get(ctx, tt.key, &result)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the value was retrieved correctly
				if result == nil {
					t.Error("Get() returned nil value")
				}
			}
		})
	}
}

func TestMemoryStore_GetNonExistentKey(t *testing.T) {
	ctx := context.Background()
	store := &MemoryStore{Data: make(map[string][]byte)}

	var result string
	err := store.Get(ctx, "non-existent", &result)
	if err != ErrKeyNotFound {
		t.Errorf("Get() error = %v, want %v", err, ErrKeyNotFound)
	}
}

func TestMemoryStore_OverwriteKey(t *testing.T) {
	ctx := context.Background()
	store := &MemoryStore{Data: make(map[string][]byte)}

	key := "test-key"

	// Set initial value
	err := store.Set(ctx, key, "initial", time.Minute)
	if err != nil {
		t.Fatalf("Set() initial error = %v", err)
	}

	// Overwrite with new value
	err = store.Set(ctx, key, "updated", time.Minute)
	if err != nil {
		t.Fatalf("Set() overwrite error = %v", err)
	}

	// Verify new value
	var result string
	err = store.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if result != "updated" {
		t.Errorf("Get() = %v, want %v", result, "updated")
	}
}

func TestMemoryStore_KeyPrefix(t *testing.T) {
	ctx := context.Background()
	store := &MemoryStore{Data: make(map[string][]byte)}

	key := "test"
	value := "data"

	err := store.Set(ctx, key, value, time.Minute)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Verify the key is stored with the prefix
	prefixedKey := dbPrefix + key
	if _, ok := store.Data[prefixedKey]; !ok {
		t.Errorf("Expected key %q to be stored with prefix, but not found", prefixedKey)
	}
}
