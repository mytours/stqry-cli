package api

import (
	"fmt"
	"strconv"
)

// PaginationMeta holds pagination information returned by list endpoints.
type PaginationMeta struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
	Count   int `json:"count"`
}

// ListCollections returns a paginated list of collections.
func ListCollections(c *Client, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		Collections []map[string]interface{} `json:"collections"`
		Meta        *PaginationMeta          `json:"meta"`
	}
	if err := c.Get("/api/public/collections", query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.Collections, resp.Meta, nil
}

// GetCollection returns a single collection by ID.
func GetCollection(c *Client, id string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Get(fmt.Sprintf("/api/public/collections/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	if col, ok := resp["collection"].(map[string]interface{}); ok {
		return col, nil
	}
	return resp, nil
}

// CreateCollection creates a new collection. Fields are sent flat; see
// CreateScreen for why.
func CreateCollection(c *Client, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Post("/api/public/collections", fields, &resp); err != nil {
		return nil, err
	}
	if col, ok := resp["collection"].(map[string]interface{}); ok {
		return col, nil
	}
	return resp, nil
}

// UpdateCollection updates an existing collection.
func UpdateCollection(c *Client, id string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Patch(fmt.Sprintf("/api/public/collections/%s", id), fields, &resp); err != nil {
		return nil, err
	}
	if col, ok := resp["collection"].(map[string]interface{}); ok {
		return col, nil
	}
	return resp, nil
}

// DeleteCollection deletes a collection by ID.
func DeleteCollection(c *Client, id string) error {
	return c.Delete(fmt.Sprintf("/api/public/collections/%s", id), nil)
}

// ListCollectionItems returns the items for a given collection.
func ListCollectionItems(c *Client, collectionID string, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		CollectionItems []map[string]interface{} `json:"collection_items"`
		Meta            *PaginationMeta          `json:"meta"`
	}
	path := fmt.Sprintf("/api/public/collections/%s/collection_items", collectionID)
	if err := c.Get(path, query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.CollectionItems, resp.Meta, nil
}

// GetCollectionItem returns a single collection item by ID.
func GetCollectionItem(c *Client, collectionID, itemID string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	path := fmt.Sprintf("/api/public/collections/%s/collection_items/%s", collectionID, itemID)
	if err := c.Get(path, nil, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp["collection_item"].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// CreateCollectionItem adds an item to a collection.
func CreateCollectionItem(c *Client, collectionID string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	path := fmt.Sprintf("/api/public/collections/%s/collection_items", collectionID)
	if err := c.Post(path, fields, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp["collection_item"].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// UpdateCollectionItem updates a collection item by ID.
func UpdateCollectionItem(c *Client, collectionID, itemID string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	path := fmt.Sprintf("/api/public/collections/%s/collection_items/%s", collectionID, itemID)
	if err := c.Patch(path, fields, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp["collection_item"].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// DeleteCollectionItem removes an item from a collection.
func DeleteCollectionItem(c *Client, collectionID, itemID string) error {
	path := fmt.Sprintf("/api/public/collections/%s/collection_items/%s", collectionID, itemID)
	return c.Delete(path, nil)
}

// ReorderCollectionItems sets the order of collection items by their IDs.
// The API expects {"positions": [{"id": <int>, "position": <int>}, ...]}.
//
// Positions are 1-based. The API treats `position: 0` as unset and clamps
// it to 1, which collided the first argument with the second (both landed
// at position 1) when we sent a 0-based sequence.
func ReorderCollectionItems(c *Client, collectionID string, itemIDs []string) error {
	path := fmt.Sprintf("/api/public/collections/%s/collection_items/update_positions", collectionID)
	positions := make([]map[string]interface{}, 0, len(itemIDs))
	for i, idStr := range itemIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid collection item id %q: must be an integer", idStr)
		}
		positions = append(positions, map[string]interface{}{"id": id, "position": i + 1})
	}
	body := map[string]interface{}{"positions": positions}
	return c.Post(path, body, nil)
}
