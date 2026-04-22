package api

import (
	"fmt"
	"strconv"
)

// ── Screens ──────────────────────────────────────────────────────────────────

// ListScreens returns a paginated list of screens.
func ListScreens(c *Client, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		Screens []map[string]interface{} `json:"screens"`
		Meta    *PaginationMeta          `json:"meta"`
	}
	if err := c.Get("/api/public/screens", query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.Screens, resp.Meta, nil
}

// GetScreen returns a single screen by ID.
func GetScreen(c *Client, id string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Get(fmt.Sprintf("/api/public/screens/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	if screen, ok := resp["screen"].(map[string]interface{}); ok {
		return screen, nil
	}
	return resp, nil
}

// CreateScreen creates a new screen. Fields are sent flat at the top level
// because the public API controller reads params.permit(:name, ...) and
// params[:type] directly; see app/controllers/api/public/screens_controller.rb
// in mytours-web.
func CreateScreen(c *Client, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Post("/api/public/screens", fields, &resp); err != nil {
		return nil, err
	}
	if screen, ok := resp["screen"].(map[string]interface{}); ok {
		return screen, nil
	}
	return resp, nil
}

// UpdateScreen updates an existing screen. See CreateScreen for why the body
// is flat.
func UpdateScreen(c *Client, id string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	if err := c.Patch(fmt.Sprintf("/api/public/screens/%s", id), fields, &resp); err != nil {
		return nil, err
	}
	if screen, ok := resp["screen"].(map[string]interface{}); ok {
		return screen, nil
	}
	return resp, nil
}

// DeleteScreen deletes a screen by ID.
func DeleteScreen(c *Client, id string) error {
	return c.Delete(fmt.Sprintf("/api/public/screens/%s", id), nil)
}

// ── Story Sections ────────────────────────────────────────────────────────────

// ListStorySections returns all story sections for a screen.
func ListStorySections(c *Client, screenID string, query map[string]string) ([]map[string]interface{}, *PaginationMeta, error) {
	var resp struct {
		StorySections []map[string]interface{} `json:"story_sections"`
		Meta          *PaginationMeta          `json:"meta"`
	}
	path := fmt.Sprintf("/api/public/screens/%s/story_sections", screenID)
	if err := c.Get(path, query, &resp); err != nil {
		return nil, nil, err
	}
	return resp.StorySections, resp.Meta, nil
}

// GetStorySection returns a single story section.
func GetStorySection(c *Client, screenID, sectionID string) (map[string]interface{}, error) {
	var resp map[string]interface{}
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/%s", screenID, sectionID)
	if err := c.Get(path, nil, &resp); err != nil {
		return nil, err
	}
	if section, ok := resp["story_section"].(map[string]interface{}); ok {
		return section, nil
	}
	return resp, nil
}

// CreateStorySection creates a new story section for a screen.
func CreateStorySection(c *Client, screenID string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	path := fmt.Sprintf("/api/public/screens/%s/story_sections", screenID)
	if err := c.Post(path, fields, &resp); err != nil {
		return nil, err
	}
	if section, ok := resp["story_section"].(map[string]interface{}); ok {
		return section, nil
	}
	return resp, nil
}

// UpdateStorySection updates an existing story section.
func UpdateStorySection(c *Client, screenID, sectionID string, fields map[string]interface{}) (map[string]interface{}, error) {
	var resp map[string]interface{}
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/%s", screenID, sectionID)
	if err := c.Patch(path, fields, &resp); err != nil {
		return nil, err
	}
	if section, ok := resp["story_section"].(map[string]interface{}); ok {
		return section, nil
	}
	return resp, nil
}

// DeleteStorySection deletes a story section.
func DeleteStorySection(c *Client, screenID, sectionID string) error {
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/%s", screenID, sectionID)
	return c.Delete(path, nil)
}

// ReorderStorySections sets the order of story sections via update_positions.
// The API expects {"positions": [{"id": <int>, "position": <int>}, ...]}.
//
// Positions are 1-based; see ReorderCollectionItems for why.
func ReorderStorySections(c *Client, screenID string, sectionIDs []string) error {
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/update_positions", screenID)
	positions := make([]map[string]interface{}, 0, len(sectionIDs))
	for i, idStr := range sectionIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid section id %q: must be an integer", idStr)
		}
		positions = append(positions, map[string]interface{}{"id": id, "position": i + 1})
	}
	body := map[string]interface{}{"positions": positions}
	return c.Post(path, body, nil)
}

// ── Generic Section Sub-Items ─────────────────────────────────────────────────

// ListSectionSubItems returns sub-items of a given type for a story section.
// subItemType is the plural API path segment (e.g. "badge_items").
func ListSectionSubItems(c *Client, screenID, sectionID, subItemType string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/%s/%s", screenID, sectionID, subItemType)
	var raw map[string]interface{}
	if err := c.Get(path, nil, &raw); err != nil {
		return nil, err
	}
	if arr, ok := raw[subItemType].([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result, nil
	}
	return []map[string]interface{}{}, nil
}

// CreateSectionSubItem creates a sub-item of the given type.
func CreateSectionSubItem(c *Client, screenID, sectionID, subItemType, singularKey string, fields map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/%s/%s", screenID, sectionID, subItemType)
	var resp map[string]interface{}
	if err := c.Post(path, fields, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp[singularKey].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// UpdateSectionSubItem updates an existing sub-item.
func UpdateSectionSubItem(c *Client, screenID, sectionID, subItemType, itemID, singularKey string, fields map[string]interface{}) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/%s/%s/%s", screenID, sectionID, subItemType, itemID)
	var resp map[string]interface{}
	if err := c.Patch(path, fields, &resp); err != nil {
		return nil, err
	}
	if item, ok := resp[singularKey].(map[string]interface{}); ok {
		return item, nil
	}
	return resp, nil
}

// DeleteSectionSubItem deletes a sub-item.
func DeleteSectionSubItem(c *Client, screenID, sectionID, subItemType, itemID string) error {
	path := fmt.Sprintf("/api/public/screens/%s/story_sections/%s/%s/%s", screenID, sectionID, subItemType, itemID)
	return c.Delete(path, nil)
}
