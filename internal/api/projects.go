package api

import "fmt"

// ListProjects returns a paginated list of projects.
func ListProjects(c *Client, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		Projects []map[string]interface{} `json:"projects"`
		Meta     *PaginationMeta          `json:"meta"`
	}
	if err := c.Get("/api/public/projects", query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.Projects, resp.Meta, nil
}

// GetProject returns a single project by ID.
func GetProject(c *Client, id string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Get(fmt.Sprintf("/api/public/projects/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	if proj, ok := resp["project"].(map[string]interface{}); ok {
		return proj, nil
	}
	return resp, nil
}
