package api

import "fmt"

// ListMediaItems returns a paginated list of media items.
func ListMediaItems(c *Client, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		MediaItems []map[string]interface{} `json:"media_items"`
		Meta       *PaginationMeta          `json:"meta"`
	}
	if err := c.Get("/api/public/media_items", query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.MediaItems, resp.Meta, nil
}

// GetMediaItem returns a single media item by ID.
func GetMediaItem(c *Client, id string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Get(fmt.Sprintf("/api/public/media_items/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp["media_item"].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// CreateMediaItem creates a new media item. Fields are sent flat; see
// CreateScreen for why.
func CreateMediaItem(c *Client, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Post("/api/public/media_items", fields, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp["media_item"].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// UpdateMediaItem updates an existing media item.
func UpdateMediaItem(c *Client, id string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Patch(fmt.Sprintf("/api/public/media_items/%s", id), fields, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp["media_item"].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// DeleteMediaItem deletes a media item by ID. Optional query params (e.g. "language").
func DeleteMediaItem(c *Client, id string, query map[string]string) error {
	return c.Delete(fmt.Sprintf("/api/public/media_items/%s", id), query)
}
