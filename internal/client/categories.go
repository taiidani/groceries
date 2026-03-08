package client

import (
	"context"
	"fmt"
	"net/http"
)

// CategoryDetail is the full category representation returned by the get-by-ID endpoint,
// including the items assigned to it.
type CategoryDetail struct {
	ID          int    `json:"id"`
	StoreID     int    `json:"store_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ItemCount   int    `json:"item_count"`
	Items       []Item `json:"items"`
}

// ListCategories returns all categories.
func (c *Client) ListCategories(ctx context.Context) ([]Category, error) {
	resp, err := c.do(ctx, http.MethodGet, "/api/v1/categories", nil)
	if err != nil {
		return nil, err
	}

	var categories []Category
	if err := decode(resp, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

// GetCategory returns a single category by ID, including its items.
func (c *Client) GetCategory(ctx context.Context, id int) (CategoryDetail, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/categories/%d", id), nil)
	if err != nil {
		return CategoryDetail{}, err
	}

	var category CategoryDetail
	if err := decode(resp, &category); err != nil {
		return CategoryDetail{}, err
	}

	return category, nil
}

// CreateCategory creates a new category and returns it with its assigned ID.
func (c *Client) CreateCategory(ctx context.Context, storeID int, name, description string) (Category, error) {
	body := struct {
		StoreID     int    `json:"store_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		StoreID:     storeID,
		Name:        name,
		Description: description,
	}

	resp, err := c.do(ctx, http.MethodPost, "/api/v1/categories", body)
	if err != nil {
		return Category{}, err
	}

	var category Category
	if err := decode(resp, &category); err != nil {
		return Category{}, err
	}

	return category, nil
}

// UpdateCategory updates a category and returns the updated category.
func (c *Client) UpdateCategory(ctx context.Context, id, storeID int, name, description string) (Category, error) {
	body := struct {
		StoreID     int    `json:"store_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		StoreID:     storeID,
		Name:        name,
		Description: description,
	}

	resp, err := c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/categories/%d", id), body)
	if err != nil {
		return Category{}, err
	}

	var category Category
	if err := decode(resp, &category); err != nil {
		return Category{}, err
	}

	return category, nil
}

// DeleteCategory deletes a category by ID.
func (c *Client) DeleteCategory(ctx context.Context, id int) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%d", id), nil)
	if err != nil {
		return err
	}

	return checkError(resp)
}
