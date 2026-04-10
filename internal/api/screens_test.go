package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListScreens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screens": []interface{}{
				map[string]interface{}{"id": "1", "name": "Screen One"},
				map[string]interface{}{"id": "2", "name": "Screen Two"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 2,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	screens, meta, err := ListScreens(c, nil)
	if err != nil {
		t.Fatalf("ListScreens: %v", err)
	}
	if len(screens) != 2 {
		t.Errorf("expected 2 screens, got %d", len(screens))
	}
	if screens[0]["name"] != "Screen One" {
		t.Errorf("unexpected first screen name: %v", screens[0]["name"])
	}
	if meta == nil {
		t.Error("expected non-nil meta")
	} else if meta.Count != 2 {
		t.Errorf("expected count=2, got %d", meta.Count)
	}
}

func TestListScreensWithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("q") != "museum" {
			t.Errorf("expected q=museum, got %s", r.URL.Query().Get("q"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screens": []interface{}{},
			"meta":    map[string]interface{}{"page": 2, "pages": 1, "per_page": 30, "count": 0},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	_, _, err := ListScreens(c, map[string]string{"page": "2", "q": "museum"})
	if err != nil {
		t.Fatalf("ListScreens with query: %v", err)
	}
}

func TestListStorySections(t *testing.T) {
	const screenID = "42"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/api/public/screens/" + screenID + "/story_sections"
		if r.URL.Path != expected {
			t.Errorf("expected path %s, got %s", expected, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"story_sections": []interface{}{
				map[string]interface{}{"id": "10", "section_type": "info"},
				map[string]interface{}{"id": "11", "section_type": "gallery"},
			},
			"meta": map[string]interface{}{"page": 1, "pages": 1, "per_page": 30, "count": 2},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	sections, meta, err := ListStorySections(c, screenID, nil)
	if err != nil {
		t.Fatalf("ListStorySections: %v", err)
	}
	if len(sections) != 2 {
		t.Errorf("expected 2 sections, got %d", len(sections))
	}
	if sections[1]["section_type"] != "gallery" {
		t.Errorf("unexpected section_type: %v", sections[1]["section_type"])
	}
	if meta == nil {
		t.Error("expected non-nil meta")
	}
}

func TestListSectionSubItems(t *testing.T) {
	const (
		screenID    = "42"
		sectionID   = "10"
		subItemType = "badge_items"
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/api/public/screens/" + screenID + "/story_sections/" + sectionID + "/" + subItemType
		if r.URL.Path != expected {
			t.Errorf("expected path %s, got %s", expected, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			subItemType: []interface{}{
				map[string]interface{}{"id": "100", "badge_id": "badge-abc"},
				map[string]interface{}{"id": "101", "badge_id": "badge-def"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	items, err := ListSectionSubItems(c, screenID, sectionID, subItemType)
	if err != nil {
		t.Fatalf("ListSectionSubItems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 badge_items, got %d", len(items))
	}
	if items[0]["badge_id"] != "badge-abc" {
		t.Errorf("unexpected badge_id: %v", items[0]["badge_id"])
	}
}

func TestListSectionSubItemsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"badge_items": []interface{}{},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	items, err := ListSectionSubItems(c, "1", "2", "badge_items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestGetScreen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "42", "name": "Museum Entry"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	screen, err := GetScreen(c, "42")
	if err != nil {
		t.Fatalf("GetScreen: %v", err)
	}
	if screen["name"] != "Museum Entry" {
		t.Errorf("expected name=Museum Entry, got %v", screen["name"])
	}
}

func TestCreateScreen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		// The public API expects fields flat at the top level, not wrapped
		// under a "screen" key. See app/controllers/api/public/screens_controller.rb.
		if _, wrapped := body["screen"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"screen\": %v", body)
		}
		if body["name"] != "New Screen" {
			t.Errorf("expected name=New Screen, got %v", body["name"])
		}
		if body["type"] != "story" {
			t.Errorf("expected type=story, got %v", body["type"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "99", "name": "New Screen", "type": "story"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	screen, err := CreateScreen(c, map[string]interface{}{"name": "New Screen", "type": "story"})
	if err != nil {
		t.Fatalf("CreateScreen: %v", err)
	}
	if screen["name"] != "New Screen" {
		t.Errorf("expected name=New Screen, got %v", screen["name"])
	}
}

func TestUpdateScreen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["screen"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"screen\": %v", body)
		}
		if body["name"] != "Updated" {
			t.Errorf("expected name=Updated, got %v", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "42", "name": "Updated"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	screen, err := UpdateScreen(c, "42", map[string]interface{}{"name": "Updated"})
	if err != nil {
		t.Fatalf("UpdateScreen: %v", err)
	}
	if screen["name"] != "Updated" {
		t.Errorf("expected name=Updated, got %v", screen["name"])
	}
}

func TestDeleteScreen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := DeleteScreen(c, "42")
	if err != nil {
		t.Fatalf("DeleteScreen: %v", err)
	}
}

func TestGetStorySection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections/10" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"story_section": map[string]interface{}{"id": "10", "section_type": "info"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	section, err := GetStorySection(c, "42", "10")
	if err != nil {
		t.Fatalf("GetStorySection: %v", err)
	}
	if section["section_type"] != "info" {
		t.Errorf("expected section_type=info, got %v", section["section_type"])
	}
}

func TestCreateStorySection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["story_section"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"story_section\": %v", body)
		}
		if body["type"] != "media_group" {
			t.Errorf("expected type=media_group, got %v", body["type"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"story_section": map[string]interface{}{"id": "12", "type": "media_group"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	section, err := CreateStorySection(c, "42", map[string]interface{}{"type": "media_group"})
	if err != nil {
		t.Fatalf("CreateStorySection: %v", err)
	}
	if section["type"] != "media_group" {
		t.Errorf("expected type=media_group, got %v", section["type"])
	}
}

func TestUpdateStorySection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections/10" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["story_section"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"story_section\": %v", body)
		}
		if body["position"] != float64(3) {
			t.Errorf("expected position=3, got %v", body["position"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"story_section": map[string]interface{}{"id": "10", "position": 3},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	section, err := UpdateStorySection(c, "42", "10", map[string]interface{}{"position": 3})
	if err != nil {
		t.Fatalf("UpdateStorySection: %v", err)
	}
	if section["position"] != float64(3) {
		t.Errorf("expected position=3, got %v", section["position"])
	}
}

func TestDeleteStorySection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections/10" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := DeleteStorySection(c, "42", "10")
	if err != nil {
		t.Fatalf("DeleteStorySection: %v", err)
	}
}

func TestReorderStorySections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections/update_positions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		ids, ok := body["section_ids"].([]interface{})
		if !ok {
			t.Fatalf("expected body.section_ids to be a slice, got %T", body["section_ids"])
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 section_ids, got %d", len(ids))
		}
		if ids[0] != "11" || ids[1] != "10" {
			t.Errorf("unexpected section_ids order: %v", ids)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := ReorderStorySections(c, "42", []string{"11", "10"})
	if err != nil {
		t.Fatalf("ReorderStorySections: %v", err)
	}
}

func TestCreateSectionSubItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections/10/badge_items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["badge_item"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"badge_item\": %v", body)
		}
		if body["badge_id"] != "badge-xyz" {
			t.Errorf("expected badge_id=badge-xyz, got %v", body["badge_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"badge_item": map[string]interface{}{"id": "200", "badge_id": "badge-xyz"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	item, err := CreateSectionSubItem(c, "42", "10", "badge_items", "badge_item", map[string]interface{}{"badge_id": "badge-xyz"})
	if err != nil {
		t.Fatalf("CreateSectionSubItem: %v", err)
	}
	if item["badge_id"] != "badge-xyz" {
		t.Errorf("expected badge_id=badge-xyz, got %v", item["badge_id"])
	}
}

func TestUpdateSectionSubItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections/10/badge_items/100" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["badge_item"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"badge_item\": %v", body)
		}
		if body["badge_id"] != "badge-updated" {
			t.Errorf("expected badge_id=badge-updated, got %v", body["badge_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"badge_item": map[string]interface{}{"id": "100", "badge_id": "badge-updated"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	item, err := UpdateSectionSubItem(c, "42", "10", "badge_items", "100", "badge_item", map[string]interface{}{"badge_id": "badge-updated"})
	if err != nil {
		t.Fatalf("UpdateSectionSubItem: %v", err)
	}
	if item["badge_id"] != "badge-updated" {
		t.Errorf("expected badge_id=badge-updated, got %v", item["badge_id"])
	}
}

func TestDeleteSectionSubItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/screens/42/story_sections/10/badge_items/100" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := DeleteSectionSubItem(c, "42", "10", "badge_items", "100")
	if err != nil {
		t.Fatalf("DeleteSectionSubItem: %v", err)
	}
}
