package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"projects": []interface{}{
				map[string]interface{}{"id": 1, "name": "Project Alpha"},
				map[string]interface{}{"id": 2, "name": "Project Beta"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 2,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	projects, meta, err := ListProjects(c, nil)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
	if meta == nil {
		t.Error("expected non-nil meta")
	} else if meta.Count != 2 {
		t.Errorf("expected meta.Count=2, got %d", meta.Count)
	}
}

func TestGetProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/projects/10" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"project": map[string]interface{}{"id": 10, "name": "My Project"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	project, err := GetProject(c, "10")
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if project["name"] != "My Project" {
		t.Errorf("expected name=My Project, got %v", project["name"])
	}
}
