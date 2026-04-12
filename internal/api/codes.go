package api

import "fmt"

// ListCodes returns a paginated list of codes.
func ListCodes(c *Client, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		Codes []map[string]interface{} `json:"codes"`
		Meta  *PaginationMeta          `json:"meta"`
	}
	if err := c.Get("/api/public/codes", query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.Codes, resp.Meta, nil
}

// GetCode returns a single code by ID.
func GetCode(c *Client, id string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Get(fmt.Sprintf("/api/public/codes/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	if code, ok := resp["code"].(map[string]interface{}); ok {
		return code, nil
	}
	return resp, nil
}

// CreateCode creates a new code. Fields are sent flat; see CreateScreen for why.
func CreateCode(c *Client, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Post("/api/public/codes", fields, &resp); err != nil {
		return nil, err
	}
	if code, ok := resp["code"].(map[string]interface{}); ok {
		return code, nil
	}
	return resp, nil
}

// UpdateCode updates an existing code.
func UpdateCode(c *Client, id string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Patch(fmt.Sprintf("/api/public/codes/%s", id), fields, &resp); err != nil {
		return nil, err
	}
	if code, ok := resp["code"].(map[string]interface{}); ok {
		return code, nil
	}
	return resp, nil
}

// DeleteCode deletes a code by ID.
func DeleteCode(c *Client, id string) error {
	return c.Delete(fmt.Sprintf("/api/public/codes/%s", id), nil)
}
