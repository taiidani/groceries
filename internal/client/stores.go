package client

import (
	"context"
	"fmt"
	"net/http"
)

// Store mirrors the API's Store response shape. It is intentionally separate
// from models.Store so the client is not coupled to the internal model struct.
type Store struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Categories []Category `json:"categories,omitempty"`
}

// Category is the minimal category representation returned alongside a store.
type Category struct {
	ID          int    `json:"id"`
	StoreID     int    `json:"store_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ItemCount   int    `json:"item_count"`
}

// ListStores returns all stores.
func (c *Client) ListStores(ctx context.Context) ([]Store, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/stores", nil)
	if err != nil {
		return nil, err
	}

	var stores []Store
	if err := decode(resp, &stores); err != nil {
		return nil, err
	}

	return stores, nil
}

// GetStore returns a single store by ID, including its categories.
func (c *Client) GetStore(ctx context.Context, id int) (Store, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/stores/%d", id), nil)
	if err != nil {
		return Store{}, err
	}

	var store Store
	if err := decode(resp, &store); err != nil {
		return Store{}, err
	}

	return store, nil
}

// CreateStore creates a new store and returns it with its assigned ID.
func (c *Client) CreateStore(ctx context.Context, name string) (Store, error) {
	body := struct {
		Name string `json:"name"`
	}{Name: name}

	resp, err := c.do(ctx, http.MethodPost, "/api/v1/stores", body)
	if err != nil {
		return Store{}, err
	}

	var store Store
	if err := decode(resp, &store); err != nil {
		return Store{}, err
	}

	return store, nil
}

// UpdateStore updates a store's name and returns the updated store.
func (c *Client) UpdateStore(ctx context.Context, id int, name string) (Store, error) {
	body := struct {
		Name string `json:"name"`
	}{Name: name}

	resp, err := c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/stores/%d", id), body)
	if err != nil {
		return Store{}, err
	}

	var store Store
	if err := decode(resp, &store); err != nil {
		return Store{}, err
	}

	return store, nil
}

// DeleteStore deletes a store by ID.
func (c *Client) DeleteStore(ctx context.Context, id int) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/stores/%d", id), nil)
	if err != nil {
		return err
	}

	return checkError(resp)
}
