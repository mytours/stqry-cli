package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListUploadedFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/uploaded_files" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uploaded_files": []interface{}{
				map[string]interface{}{
					"id": 1, "filename": "photo.jpg", "content_type": "image/jpeg", "file_size": 12345,
				},
			},
			"meta": map[string]interface{}{"page": 1, "pages": 1, "per_page": 30, "count": 1},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	files, meta, err := ListUploadedFiles(c, nil)
	if err != nil {
		t.Fatalf("ListUploadedFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 row, got %d", len(files))
	}
	if files[0]["filename"] != "photo.jpg" {
		t.Errorf("expected filename=photo.jpg, got %v", files[0]["filename"])
	}
	if meta == nil || meta.Count != 1 {
		t.Errorf("expected meta.Count=1, got %+v", meta)
	}
}

func TestGetUploadedFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/uploaded_files/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uploaded_file": map[string]interface{}{
				"id": 42, "filename": "audio.mp3", "content_type": "audio/mp3",
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	got, err := GetUploadedFile(c, "42")
	if err != nil {
		t.Fatalf("GetUploadedFile: %v", err)
	}
	if got["filename"] != "audio.mp3" {
		t.Errorf("expected filename=audio.mp3, got %v", got["filename"])
	}
}
