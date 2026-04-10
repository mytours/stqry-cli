package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMediaItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/media_items/mi-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "mi-1", "name": "Video 1"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	item, err := GetMediaItem(c, "mi-1")
	if err != nil {
		t.Fatalf("GetMediaItem: %v", err)
	}
	if item["name"] != "Video 1" {
		t.Errorf("expected name=Video 1, got %v", item["name"])
	}
}

func TestCreateMediaItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/media_items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["media_item"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"media_item\": %v", body)
		}
		if body["name"] != "New Video" {
			t.Errorf("expected name=New Video, got %v", body["name"])
		}
		if body["type"] != "video" {
			t.Errorf("expected type=video, got %v", body["type"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "mi-2", "name": "New Video", "type": "video"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	item, err := CreateMediaItem(c, map[string]interface{}{"name": "New Video", "type": "video"})
	if err != nil {
		t.Fatalf("CreateMediaItem: %v", err)
	}
	if item["name"] != "New Video" {
		t.Errorf("expected name=New Video, got %v", item["name"])
	}
}

func TestUpdateMediaItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/media_items/mi-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["media_item"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"media_item\": %v", body)
		}
		if body["name"] != "Updated" {
			t.Errorf("expected name=Updated, got %v", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "mi-1", "name": "Updated"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	item, err := UpdateMediaItem(c, "mi-1", map[string]interface{}{"name": "Updated"})
	if err != nil {
		t.Fatalf("UpdateMediaItem: %v", err)
	}
	if item["name"] != "Updated" {
		t.Errorf("expected name=Updated, got %v", item["name"])
	}
}

func TestDeleteMediaItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/media_items/mi-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := DeleteMediaItem(c, "mi-1", nil)
	if err != nil {
		t.Fatalf("DeleteMediaItem: %v", err)
	}
}

func TestDeleteMediaItemWithLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/media_items/mi-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("language") != "fr" {
			t.Errorf("expected language=fr, got %s", r.URL.Query().Get("language"))
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	err := DeleteMediaItem(c, "mi-1", map[string]string{"language": "fr"})
	if err != nil {
		t.Fatalf("DeleteMediaItem with language: %v", err)
	}
}
