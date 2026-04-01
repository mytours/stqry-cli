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
