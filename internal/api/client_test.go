package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Token") != "test-token" {
			t.Errorf("expected X-Api-Token: test-token, got %s", r.Header.Get("X-Api-Token"))
		}
		if r.URL.Path != "/api/public/collections" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collections": []interface{}{},
			"meta":        map[string]interface{}{"page": 1, "pages": 1, "per_page": 30, "count": 0},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	var result map[string]interface{}
	err := c.Get("/api/public/collections", nil, &result)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if result["collections"] == nil {
		t.Error("expected collections key")
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test" {
			t.Errorf("expected name=test, got %v", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1, "name": "test"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	var result map[string]interface{}
	err := c.Post("/api/public/collections", map[string]interface{}{"name": "test"}, &result)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
}

func TestClientErrorResponse(t *testing.T) {
	// The Rails public API returns errors as [{code, message}, ...]; see
	// app/views/api/public/shared/_errors.json.jbuilder in mytours-web.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []interface{}{
				map[string]interface{}{"code": "field_validation", "message": "type: is invalid or missing"},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	var result map[string]interface{}
	err := c.Get("/api/public/collections", nil, &result)
	if err == nil {
		t.Fatal("expected error for 422 response")
	}
	if !strings.Contains(err.Error(), "type: is invalid or missing") {
		t.Errorf("expected error message to include server-side message, got %q", err.Error())
	}
}

func TestClientQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("per_page") != "10" {
			t.Errorf("expected per_page=10, got %s", r.URL.Query().Get("per_page"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	var result map[string]interface{}
	c.Get("/api/public/collections", map[string]string{"page": "2", "per_page": "10"}, &result)
}
