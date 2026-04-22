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
		// Rails public API expects flat params; see collections_controller.rb.
		if _, wrapped := body["collection"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"collection\": %v", body)
		}
		if body["name"] != "new-col" {
			t.Errorf("expected name=new-col, got %v", body["name"])
		}
		if body["type"] != "list" {
			t.Errorf("expected type=list, got %v", body["type"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 99, "name": "new-col", "type": "list"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	col, err := CreateCollection(c, map[string]interface{}{"name": "new-col", "type": "list"})
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

func TestUpdateCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["collection"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"collection\": %v", body)
		}
		if body["name"] != "updated-col" {
			t.Errorf("expected name=updated-col, got %v", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 42, "name": "updated-col"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	col, err := UpdateCollection(c, "42", map[string]interface{}{"name": "updated-col"})
	if err != nil {
		t.Fatalf("UpdateCollection: %v", err)
	}
	if col["name"] != "updated-col" {
		t.Errorf("expected name=updated-col, got %v", col["name"])
	}
}

func TestDeleteCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := DeleteCollection(c, "42")
	if err != nil {
		t.Fatalf("DeleteCollection: %v", err)
	}
}

func TestCreateCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections/7/collection_items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["collection_item"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"collection_item\": %v", body)
		}
		if body["item_type"] != "Screen" {
			t.Errorf("expected item_type=Screen, got %v", body["item_type"])
		}
		if body["item_id"] != "10" {
			t.Errorf("expected item_id=10, got %v", body["item_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection_item": map[string]interface{}{"id": 3, "item_type": "Screen", "item_id": "10"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	item, err := CreateCollectionItem(c, "7", map[string]interface{}{"item_type": "Screen", "item_id": "10"})
	if err != nil {
		t.Fatalf("CreateCollectionItem: %v", err)
	}
	if item["item_type"] != "Screen" {
		t.Errorf("expected item_type=Screen, got %v", item["item_type"])
	}
}

func TestUpdateCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections/7/collection_items/1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["collection_item"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"collection_item\": %v", body)
		}
		if body["position"] != float64(3) {
			t.Errorf("expected position=3, got %v", body["position"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection_item": map[string]interface{}{"id": 1, "position": 3},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	item, err := UpdateCollectionItem(c, "7", "1", map[string]interface{}{"position": 3})
	if err != nil {
		t.Fatalf("UpdateCollectionItem: %v", err)
	}
	if item["position"] != float64(3) {
		t.Errorf("expected position=3, got %v", item["position"])
	}
}

func TestDeleteCollectionItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections/7/collection_items/1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := DeleteCollectionItem(c, "7", "1")
	if err != nil {
		t.Fatalf("DeleteCollectionItem: %v", err)
	}
}

func TestReorderCollectionItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/collections/7/collection_items/update_positions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		positions, ok := body["positions"].([]interface{})
		if !ok {
			t.Fatalf("expected body.positions to be a slice, got %T", body["positions"])
		}
		if len(positions) != 2 {
			t.Errorf("expected 2 positions, got %d", len(positions))
		}
		first := positions[0].(map[string]interface{})
		second := positions[1].(map[string]interface{})
		// Positions are 1-based: the API treats position 0 as unset
		// and clamps it to 1, which would collide the first two items.
		if first["id"].(float64) != 2 || first["position"].(float64) != 1 {
			t.Errorf("unexpected first position: %v", first)
		}
		if second["id"].(float64) != 1 || second["position"].(float64) != 2 {
			t.Errorf("unexpected second position: %v", second)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := ReorderCollectionItems(c, "7", []string{"2", "1"})
	if err != nil {
		t.Fatalf("ReorderCollectionItems: %v", err)
	}
}
