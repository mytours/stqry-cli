package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCollections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %s", r.URL.Query().Get("page"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collections": []interface{}{
				map[string]interface{}{"id": 1, "name": "alpha"},
				map[string]interface{}{"id": 2, "name": "beta"},
			},
			"meta": map[string]interface{}{
				"page": 2, "pages": 5, "per_page": 30, "count": 120,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	cols, meta, err := ListCollections(c, map[string]string{"page": "2"})
	if err != nil {
		t.Fatalf("ListCollections: %v", err)
	}
	if len(cols) != 2 {
		t.Errorf("expected 2 collections, got %d", len(cols))
	}
	if meta.Page != 2 {
		t.Errorf("expected meta.Page=2, got %d", meta.Page)
	}
	if meta.Pages != 5 {
		t.Errorf("expected meta.Pages=5, got %d", meta.Pages)
	}
}

func TestGetCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 42, "name": "my-col"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	col, err := GetCollection(c, "42")
	if err != nil {
		t.Fatalf("GetCollection: %v", err)
	}
	if col["name"] != "my-col" {
		t.Errorf("expected name=my-col, got %v", col["name"])
	}
}

func TestCreateCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		colFields, ok := body["collection"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected body.collection to be a map, got %T", body["collection"])
		}
		if colFields["name"] != "new-col" {
			t.Errorf("expected collection.name=new-col, got %v", colFields["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 99, "name": "new-col"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	col, err := CreateCollection(c, map[string]interface{}{"name": "new-col"})
	if err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}
	if col["name"] != "new-col" {
		t.Errorf("expected name=new-col, got %v", col["name"])
	}
}

func TestListCollectionItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		expectedPath := "/api/public/collections/7/collection_items"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection_items": []interface{}{
				map[string]interface{}{"id": 1, "item_type": "Screen", "item_id": 10, "position": 1},
				map[string]interface{}{"id": 2, "item_type": "Screen", "item_id": 20, "position": 2},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 2,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	items, meta, err := ListCollectionItems(c, "7", nil)
	if err != nil {
		t.Fatalf("ListCollectionItems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if meta.Count != 2 {
		t.Errorf("expected meta.Count=2, got %d", meta.Count)
	}
}
