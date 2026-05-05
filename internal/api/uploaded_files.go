package api

import "fmt"

// ListUploadedFiles returns a paginated list of uploaded_file records — the
// underlying binary metadata (filename, size, content_type, hash, dimensions,
// duration) referenced by media_items.
func ListUploadedFiles(c *Client, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		UploadedFiles []map[string]interface{} `json:"uploaded_files"`
		Meta          *PaginationMeta          `json:"meta"`
	}
	if err := c.Get("/api/public/uploaded_files", query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.UploadedFiles, resp.Meta, nil
}

// GetUploadedFile returns a single uploaded_file by ID.
func GetUploadedFile(c *Client, id string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Get(fmt.Sprintf("/api/public/uploaded_files/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	if u, ok := resp["uploaded_file"].(map[string]interface{}); ok {
		return u, nil
	}
	return resp, nil
}
