package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

	// Step 3: enqueue — the Rails job expects s3_url to be the raw S3 object
	// key (e.g. "uploads/test-video.mp4"), not a full URL. See
	// spec/sidekiq/uploaded_file_process_spec.rb in mytours-web.
	var enqueuedS3URL string
	mux.HandleFunc("/api/public/uploaded_files/process_enqueue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("enqueue: expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if s, ok := body["s3_url"].(string); ok {
			enqueuedS3URL = s
		} else {
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
			"message":      "File processed successfully",
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
	var statusMessages []string
	result, err := UploadFile(c, filePath, server.URL, func(written, total int64) {
		lastWritten = written
		lastTotal = total
	}, func(msg string) {
		statusMessages = append(statusMessages, msg)
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

	// The enqueue request must carry the S3 object key, not a URL.
	if enqueuedS3URL != "uploads/test-video.mp4" {
		t.Errorf("expected enqueue s3_url to be the S3 key \"uploads/test-video.mp4\", got %q", enqueuedS3URL)
	}

	// Progress callback should have been called.
	if lastTotal == 0 {
		t.Error("expected onProgress to be called")
	}
	_ = lastWritten

	// onStatus callback should have been called with the server message.
	if len(statusMessages) != 1 || statusMessages[0] != "File processed successfully" {
		t.Errorf("expected onStatus called with processing message, got %v", statusMessages)
	}
}

// TestUploadFileSetsContentLength verifies that the S3 POST sends an explicit
// Content-Length instead of falling back to chunked transfer encoding. Real
// AWS S3 rejects chunked POST uploads with 411 Length Required; httptest
// accepts both, so this test asserts the request shape directly.
func TestUploadFileSetsContentLength(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(filePath, []byte("fake png bytes"), 0600); err != nil {
		t.Fatalf("creating temp file: %v", err)
	}

	var gotContentLength int64 = -1
	var gotTransferEncoding []string

	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/uploaded_files/presigned", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":    "",
			"fields": map[string]string{"key": "uploads/test.png"},
		})
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		gotContentLength = r.ContentLength
		gotTransferEncoding = r.TransferEncoding
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/public/uploaded_files/process_enqueue", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_id": "job-cl"})
	})
	mux.HandleFunc("/api/public/uploaded_files/process_status/job-cl", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "complete",
			"uploaded_file": map[string]interface{}{"id": "uf-1"},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	// Pass a non-nil onProgress to exercise the progressReader wrapping path.
	_, err := UploadFile(c, filePath, server.URL, func(written, total int64) {}, nil)
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	if gotContentLength <= 0 {
		t.Errorf("expected S3 upload to have Content-Length > 0, got %d", gotContentLength)
	}
	for _, enc := range gotTransferEncoding {
		if enc == "chunked" {
			t.Errorf("S3 upload must not use chunked transfer encoding (S3 rejects it with 411); got TransferEncoding=%v", gotTransferEncoding)
		}
	}
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
	_, err := UploadFile(c, filePath, server.URL, nil, nil)
	if err == nil {
		t.Fatal("expected error for transcoder_invalid_file status")
	}
}

// TestUploadFileSlowS3 verifies that the S3 upload step uses an HTTP client
// with a longer timeout than the default 30-second Client.Timeout set on
// Client.HTTPClient. Real users hit "Client.Timeout exceeded while awaiting
// headers" on medium-size (~6 MB) files when the 30 s budget covers the whole
// request lifecycle (body write + S3 commit + response headers).
func TestUploadFileSlowS3(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "slow.png")
	if err := os.WriteFile(filePath, []byte("fake png bytes"), 0600); err != nil {
		t.Fatalf("creating temp file: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/uploaded_files/presigned", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":    "",
			"fields": map[string]string{"key": "upload_file_cache/slow.png"},
		})
	})
	// Sleep longer than the main client's tightened timeout so that if
	// UploadFile reuses c.HTTPClient for the upload step, the request will
	// fail with a Client.Timeout error instead of succeeding.
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/public/uploaded_files/process_enqueue", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_id": "job-slow"})
	})
	mux.HandleFunc("/api/public/uploaded_files/process_status/job-slow", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "complete",
			"uploaded_file": map[string]interface{}{"id": "uf-slow"},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	// Tighten the main client to a budget much smaller than the /upload
	// handler's sleep. Presign/enqueue/status handlers respond instantly on
	// localhost so they still fit comfortably. Use a 10x margin (50ms vs 500ms)
	// to reduce flakiness under heavy CI load.
	c.HTTPClient.Timeout = 50 * time.Millisecond

	result, err := UploadFile(c, filePath, server.URL, nil, nil)
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if result["id"] != "uf-slow" {
		t.Errorf("expected id=uf-slow, got %v", result["id"])
	}
}

// TestUploadFileSidekiqFailed verifies that a Sidekiq::Status "failed" status
// from process_status propagates as an error instead of silently polling until
// the 5-minute timeout.
func TestUploadFileSidekiqFailed(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "bad.png")
	if err := os.WriteFile(filePath, []byte("x"), 0600); err != nil {
		t.Fatalf("creating temp file: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/uploaded_files/presigned", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":    "",
			"fields": map[string]string{"key": "upload_file_cache/123_bad.png"},
		})
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/public/uploaded_files/process_enqueue", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"job_id": "job-failed"})
	})
	mux.HandleFunc("/api/public/uploaded_files/process_status/job-failed", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": "S3Utils::NotFound: not found",
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	// Use a short deadline so the test doesn't take forever if the fix regresses.
	done := make(chan error, 1)
	go func() {
		_, err := UploadFile(c, filePath, server.URL, nil, nil)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error for failed status")
		}
		if !strings.Contains(err.Error(), "failed") && !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected error to mention the failure, got %q", err.Error())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("UploadFile did not return within 10s; likely still polling on an unknown status")
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
