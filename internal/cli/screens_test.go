package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestScreensCreateInvalidType asserts that `stqry screens create` rejects an
// unknown --type value client-side with a helpful message, before ever hitting
// the API.
func TestScreensCreateInvalidType(t *testing.T) {
	// No server needed: validation should short-circuit.
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "create", "--name=Welcome", "--type=standard"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid screen type, got nil")
	}
	if !contains(err.Error(), "invalid screen type") {
		t.Errorf("expected error to mention \"invalid screen type\", got %q", err.Error())
	}
	if !contains(err.Error(), "story") {
		t.Errorf("expected error to list valid types (story), got %q", err.Error())
	}
}

// TestScreensCreateFlatBody asserts that a valid create actually sends fields
// at the top level (flat), with `type` (not `screen_type`), matching what the
// Rails public API expects.
func TestScreensCreateFlatBody(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/screens" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "1", "name": "Welcome", "type": "story"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "create", "--name=Welcome", "--type=story"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if _, wrapped := captured["screen"]; wrapped {
		t.Errorf("expected flat body, got wrapper: %v", captured)
	}
	if captured["name"] != "Welcome" {
		t.Errorf("expected name=Welcome, got %v", captured["name"])
	}
	if captured["type"] != "story" {
		t.Errorf("expected type=story, got %v", captured["type"])
	}
	if _, legacy := captured["screen_type"]; legacy {
		t.Error("CLI still sending legacy screen_type field")
	}
}

// TestCollectionsCreateInvalidType asserts client-side validation of
// --type on `stqry collections create`.
func TestCollectionsCreateInvalidType(t *testing.T) {
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "create", "--name=test", "--type=bogus"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid collection type, got nil")
	}
	if !contains(err.Error(), "invalid collection type") {
		t.Errorf("expected error to mention \"invalid collection type\", got %q", err.Error())
	}
	if !contains(err.Error(), "tour") {
		t.Errorf("expected error to list valid types (tour), got %q", err.Error())
	}
}

// TestMediaCreateInvalidType asserts client-side validation of --type on
// `stqry media create`.
func TestMediaCreateInvalidType(t *testing.T) {
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "create", "--type=bogus"})
	cmd.SetOut(os.Stderr)
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
