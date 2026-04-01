package api

import "fmt"

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

// CreateCollection creates a new collection.
func CreateCollection(c *Client, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Post("/api/public/collections", map[string]interface{}{"collection": fields}, &resp); err != nil {
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
	if err := c.Patch(fmt.Sprintf("/api/public/collections/%s", id), map[string]interface{}{"collection": fields}, &resp); err != nil {
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

// CreateCollectionItem adds an item to a collection.
func CreateCollectionItem(c *Client, collectionID string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	path := fmt.Sprintf("/api/public/collections/%s/collection_items", collectionID)
	if err := c.Post(path, map[string]interface{}{"collection_item": fields}, &resp); err != nil {
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
	if err := c.Patch(path, map[string]interface{}{"collection_item": fields}, &resp); err != nil {
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
func ReorderCollectionItems(c *Client, collectionID string, itemIDs []string) error {
	path := fmt.Sprintf("/api/public/collections/%s/collection_items/update_positions", collectionID)
	body := map[string]interface{}{"ids": itemIDs}
	return c.Post(path, body, nil)
}
