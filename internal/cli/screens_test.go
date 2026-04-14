package cli

import (
	"encoding/json"
	"io"
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

// TestScreensListCmd asserts that `stqry screens list` prints the screen name
// returned by the API.
func TestScreensListCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/screens" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screens": []interface{}{
				map[string]interface{}{"id": "1", "name": "welcome", "type": "story"},
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
	cmd.SetArgs([]string{"--site=testsite", "screens", "list"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "welcome") {
		t.Errorf("expected output to contain %q, got:\n%s", "welcome", out)
	}
}

// TestScreensGetCmd asserts that `stqry screens get 42` prints the screen name
// returned by the API.
func TestScreensGetCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "42", "name": "welcome", "type": "story"},
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
	cmd.SetArgs([]string{"--site=testsite", "screens", "get", "42"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "welcome") {
		t.Errorf("expected output to contain %q, got:\n%s", "welcome", out)
	}
}

// TestScreensUpdateCmd asserts that `stqry screens update 42 --name=new-name`
// sends the correct field in the PATCH request body.
func TestScreensUpdateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "42", "name": "new-name", "type": "story"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "update", "42", "--name=new-name"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["name"] != "new-name" {
		t.Errorf("expected name=%q in request body, got %v", "new-name", captured["name"])
	}
}

// TestScreensDeleteCmd asserts that `stqry screens delete 42` sends a DELETE
// request to the correct endpoint.
func TestScreensDeleteCmd(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		} else {
			deleted = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "delete", "42"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !deleted {
		t.Error("expected DELETE request to have been made, but it was not")
	}
}
