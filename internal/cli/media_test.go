package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestMediaCreateInvalidType asserts client-side validation of --type on
// `stqry media create`.
func TestMediaCreateInvalidType(t *testing.T) {
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "create", "--type=bogus"})
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid media type, got nil")
	}
	if !contains(err.Error(), "invalid media type") {
		t.Errorf("expected error to mention \"invalid media type\", got %q", err.Error())
	}
	if !contains(err.Error(), "image") {
		t.Errorf("expected error to list valid types (image), got %q", err.Error())
	}
}

// TestMediaListCmd asserts that `stqry media list` prints the media item name
// returned by the API.
func TestMediaListCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/media_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_items": []interface{}{
				map[string]interface{}{"id": "1", "name": "banner", "type": "image"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 1,
			},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "list"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "banner") {
		t.Errorf("expected output to contain %q, got:\n%s", "banner", out)
	}
}

// TestMediaGetCmd asserts that `stqry media get 55` prints the media item name
// returned by the API.
func TestMediaGetCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/media_items/55" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "55", "name": "banner", "type": "image"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "get", "55"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "banner") {
		t.Errorf("expected output to contain %q, got:\n%s", "banner", out)
	}
}

// TestMediaCreateCmd asserts that `stqry media create --type=audio --name="City Tour"`
// sends the correct fields in the POST request body.
func TestMediaCreateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/media_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "99", "name": "City Tour", "type": "audio"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "create", "--type=audio", "--name=City Tour"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["type"] != "audio" {
		t.Errorf("expected type=%q in request body, got %v", "audio", captured["type"])
	}
	if captured["name"] != "City Tour" {
		t.Errorf("expected name=%q in request body, got %v", "City Tour", captured["name"])
	}
}

// TestMediaUpdateCmd asserts that `stqry media update 55 --name="New Banner"`
// sends the correct field in the PATCH request body.
func TestMediaUpdateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/public/media_items/55" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "55", "name": "New Banner", "type": "image"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "update", "55", "--name=New Banner"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["name"] != "New Banner" {
		t.Errorf("expected name=%q in request body, got %v", "New Banner", captured["name"])
	}
}

// TestMediaDeleteCmd asserts that `stqry media delete 55` sends a DELETE
// request to the correct endpoint and prints confirmation.
func TestMediaDeleteCmd(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/public/media_items/55" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		} else {
			deleted = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "delete", "55"})
	cmd.SetErr(os.Stderr)
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !deleted {
		t.Error("expected DELETE request to have been made, but it was not")
	}
	if !contains(out, "55") {
		t.Errorf("expected output to contain %q, got:\n%s", "55", out)
	}
}
