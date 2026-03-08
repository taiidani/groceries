package client

import (
	"context"
	"fmt"
	"net/http"
)

// ListEntry holds the list-specific state for an item that has been added to
// the shopping list.
type ListEntry struct {
	ID       int    `json:"id"`
	Quantity string `json:"quantity"`
	Done     bool   `json:"done"`
}

// Item is the full item representation returned by the items API.
// The List field is non-nil when the item is currently on the shopping list.
type Item struct {
	ID         int        `json:"id"`
	CategoryID int        `json:"category_id"`
	Name       string     `json:"name"`
	List       *ListEntry `json:"list"`
}

// ListItems returns all items. Pass inList=true to return only items currently
// on the shopping list, or inList=false to return only items not on the list.
// Pass nil to return all items regardless of list status.
func (c *Client) ListItems(ctx context.Context, inList *bool) ([]Item, error) {
	path := "/api/v1/items"
	if inList != nil {
		if *inList {
			path += "?in_list=true"
		} else {
			path += "?in_list=false"
		}
	}

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := decode(resp, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// ListShoppingList returns only the items currently on the shopping list.
func (c *Client) ListShoppingList(ctx context.Context) ([]Item, error) {
	inList := true
	return c.ListItems(ctx, &inList)
}

// GetItem returns a single item by ID.
func (c *Client) GetItem(ctx context.Context, id int) (Item, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/items/%d", id), nil)
	if err != nil {
		return Item{}, err
	}

	var item Item
	if err := decode(resp, &item); err != nil {
		return Item{}, err
	}

	return item, nil
}

// CreateItem creates a new item and returns it with its assigned ID.
func (c *Client) CreateItem(ctx context.Context, categoryID int, name string) (Item, error) {
	body := struct {
		CategoryID int    `json:"category_id"`
		Name       string `json:"name"`
	}{
		CategoryID: categoryID,
		Name:       name,
	}

	resp, err := c.do(ctx, http.MethodPost, "/api/v1/items", body)
	if err != nil {
		return Item{}, err
	}

	var item Item
	if err := decode(resp, &item); err != nil {
		return Item{}, err
	}

	return item, nil
}

// UpdateItem updates an item's category and name, and returns the updated item.
func (c *Client) UpdateItem(ctx context.Context, id, categoryID int, name string) (Item, error) {
	body := struct {
		CategoryID int    `json:"category_id"`
		Name       string `json:"name"`
	}{
		CategoryID: categoryID,
		Name:       name,
	}

	resp, err := c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/items/%d", id), body)
	if err != nil {
		return Item{}, err
	}

	var item Item
	if err := decode(resp, &item); err != nil {
		return Item{}, err
	}

	return item, nil
}

// UpdateListItem updates the quantity of an item that is currently on the
// shopping list. The id is the item ID (not the list entry ID).
func (c *Client) UpdateListItem(ctx context.Context, id int, quantity string) error {
	body := struct {
		Quantity *string `json:"quantity"`
	}{
		Quantity: &quantity,
	}

	resp, err := c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/list/items/%d", id), body)
	if err != nil {
		return err
	}

	return checkError(resp)
}

// DeleteItem deletes an item by ID.
func (c *Client) DeleteItem(ctx context.Context, id int) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/items/%d", id), nil)
	if err != nil {
		return err
	}

	return checkError(resp)
}
