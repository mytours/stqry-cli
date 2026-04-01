package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestUploadFile exercises the full upload flow against a mock HTTP server.
func TestUploadFile(t *testing.T) {
	// Create a temporary file to upload.
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-video.mp4")
	if err := os.WriteFile(filePath, []byte("fake video content"), 0600); err != nil {
		t.Fatalf("creating temp file: %v", err)
	}

	mux := http.NewServeMux()

	// Step 1: presigned
	mux.HandleFunc("/api/public/uploaded_files/presigned", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("presigned: expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":    "", // will be filled in by s3BaseURL
			"fields": map[string]string{"key": "uploads/test-video.mp4", "Content-Type": "video/mp4"},
		})
	})

	// Step 2: S3 upload endpoint (test override)
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("upload: expected POST, got %s", r.Method)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("parsing multipart: %v", err)
		}
		if r.FormValue("key") == "" {
			t.Error("expected key field in multipart form")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Step 3: enqueue
	mux.HandleFunc("/api/public/uploaded_files/process_enqueue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("enqueue: expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["s3_url"] == nil {
			t.Error("expected s3_url in enqueue body")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job_id": "job-123"})
	})

	// Step 4: process_status
	mux.HandleFunc("/api/public/uploaded_files/process_status/job-123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("process_status: expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "complete",
			"pct_complete": 100,
			"message":      "",
			"uploaded_file": map[string]interface{}{
				"id":   "uf-abc",
				"name": "test-video.mp4",
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewClient(server.URL, "test-token")

	var lastWritten, lastTotal int64
	result, err := UploadFile(c, filePath, server.URL, func(written, total int64) {
		lastWritten = written
		lastTotal = total
	})
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if result == nil {
		t.Fatal("expected uploaded_file result, got nil")
	}
	if result["id"] != "uf-abc" {
		t.Errorf("expected id=uf-abc, got %v", result["id"])
	}

	// Progress callback should have been called.
	if lastTotal == 0 {
		t.Error("expected onProgress to be called")
	}
	_ = lastWritten
}

// TestUploadFileError verifies that an error status from process_status propagates correctly.
func TestUploadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "bad.mp4")
	if err := os.WriteFile(filePath, []byte("bad content"), 0600); err != nil {
		t.Fatalf("creating temp file: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/uploaded_files/presigned", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":    "",
			"fields": map[string]string{"key": "uploads/bad.mp4"},
		})
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/public/uploaded_files/process_enqueue", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_id": "job-err"})
	})
	mux.HandleFunc("/api/public/uploaded_files/process_status/job-err", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "transcoder_invalid_file",
			"message": "unsupported format",
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	_, err := UploadFile(c, filePath, server.URL, nil)
	if err == nil {
		t.Fatal("expected error for transcoder_invalid_file status")
	}
}

// TestListMediaItems verifies the list endpoint wrapper.
func TestListMediaItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/media_items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_items": []interface{}{
				map[string]interface{}{"id": "mi-1", "name": "Video 1"},
				map[string]interface{}{"id": "mi-2", "name": "Video 2"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 2,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	items, meta, err := ListMediaItems(c, nil)
	if err != nil {
		t.Fatalf("ListMediaItems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if meta == nil {
		t.Fatal("expected meta, got nil")
	}
	if meta.Count != 2 {
		t.Errorf("expected count=2, got %d", meta.Count)
	}
}
