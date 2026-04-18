package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mytours/stqry-cli/internal/output"
)

// setupTestHome creates a temp HOME directory and writes a global config YAML
// that points the "testsite" site at the given server URL. It also initialises
// the package-level printer so commands that call printer.PrintList do not
// panic when PersistentPreRunE re-initialises it.
func setupTestHome(t *testing.T, serverURL string) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfgDir := filepath.Join(tmpDir, ".config", "stqry")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}

	cfg := fmt.Sprintf("sites:\n  testsite:\n    token: test-token\n    api_url: %s\n", serverURL)
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfg), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	// Ensure printer is non-nil so commands that reference it before
	// PersistentPreRunE runs (e.g. in error paths) don't panic.
	printer = &output.Printer{}
}

func TestCollectionsListCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/collections" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collections": []interface{}{
				map[string]interface{}{"id": 1, "name": "alpha"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 1,
			},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	// Capture stdout via os.Pipe because Printer writes directly to os.Stdout.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "list"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes := make([]byte, 4096)
	n, _ := r.Read(outBytes)
	r.Close()
	out := string(outBytes[:n])

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}

	if !contains(out, "alpha") {
		t.Errorf("expected output to contain %q, got:\n%s", "alpha", out)
	}
}

func TestJQFlagInvalidExpression(t *testing.T) {
	// Track whether the mock server receives any requests.
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "list", "--jq", "!!invalid!!"})
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for invalid jq expression")
	}
	if called {
		t.Error("API should not be called when jq expression is invalid")
	}
}

func TestJQFlagFiltersOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collections": []interface{}{
				map[string]interface{}{"id": 1, "name": "alpha"},
				map[string]interface{}{"id": 2, "name": "beta"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 2,
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
	cmd.SetArgs([]string{"--site=testsite", "collections", "list", "--jq", ".[].name"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes := make([]byte, 4096)
	n, _ := r.Read(outBytes)
	r.Close()
	out := string(outBytes[:n])

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, `"alpha"`) {
		t.Errorf("expected alpha in jq output, got:\n%s", out)
	}
	if !contains(out, `"beta"`) {
		t.Errorf("expected beta in jq output, got:\n%s", out)
	}
	// Human-readable table columns should NOT appear
	if contains(out, "NAME") {
		t.Errorf("table header should not appear in jq output, got:\n%s", out)
	}
}

// TestCollectionsCreateCmdDescription asserts that passing --description sends
// description: {<lang>: value} on the create call, so callers don't need a
// follow-up update to set a description.
func TestCollectionsCreateCmdDescription(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/collections" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 1, "name": "ct", "type": "tour"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "create",
		"--name=ct", "--type=tour", "--description=A walking tour"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	desc, ok := captured["description"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected description to be a translated-string map, got %v (type %T)", captured["description"], captured["description"])
	}
	if desc["en"] != "A walking tour" {
		t.Errorf("expected description.en=\"A walking tour\", got %v", desc["en"])
	}
}

// TestCollectionsCreateCmdOmitsDescription asserts that when --description is
// not passed, the create payload does NOT include a description field (so the
// API uses whatever default it applies).
func TestCollectionsCreateCmdOmitsDescription(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 1, "name": "ct", "type": "tour"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "create",
		"--name=ct", "--type=tour"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, present := captured["description"]; present {
		t.Errorf("expected no description field when --description is omitted, got %v", captured["description"])
	}
}

// TestCollectionsItemsAddCmdPosition asserts that passing --position sends
// position as an integer in the create body.
func TestCollectionsItemsAddCmdPosition(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/collections/42/collection_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection_item": map[string]interface{}{"id": 7, "item_type": "Screen", "item_id": 99, "position": 3},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "items", "add", "42",
		"--item-type=Screen", "--item-id=99", "--position=3"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// JSON numbers decode as float64.
	pos, ok := captured["position"].(float64)
	if !ok {
		t.Fatalf("expected position to be a number, got %v (type %T)", captured["position"], captured["position"])
	}
	if int(pos) != 3 {
		t.Errorf("expected position=3, got %v", pos)
	}
}

// TestCollectionsItemsAddCmdOmitsPosition asserts that when --position is not
// passed, the create body omits position entirely (so the API appends).
func TestCollectionsItemsAddCmdOmitsPosition(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection_item": map[string]interface{}{"id": 7, "item_type": "Screen", "item_id": 99},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "items", "add", "42",
		"--item-type=Screen", "--item-id=99"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, present := captured["position"]; present {
		t.Errorf("expected no position field when --position is omitted, got %v", captured["position"])
	}
}

// TestCollectionsCreateCmdTourType asserts that --tour-type is sent as a flat
// string field (not a TranslatedString). Tour type is an enum, not localised.
func TestCollectionsCreateCmdTourType(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 1, "name": "ct", "type": "tour", "tour_type": "walking"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "create",
		"--name=ct", "--type=tour", "--tour-type=walking"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured["tour_type"] != "walking" {
		t.Errorf("expected tour_type=\"walking\" (flat string), got %v (type %T)", captured["tour_type"], captured["tour_type"])
	}
}

// TestCollectionsCreateCmdInvalidTourType asserts client-side rejection of
// bogus tour types so callers get a fast local error with the valid list
// instead of a vague server 422.
func TestCollectionsCreateCmdInvalidTourType(t *testing.T) {
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "create",
		"--name=ct", "--type=tour", "--tour-type=spaceship"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid tour type, got nil")
	}
	if !contains(err.Error(), "invalid tour type") {
		t.Errorf("expected error to mention \"invalid tour type\", got %q", err.Error())
	}
	if !contains(err.Error(), "walking") {
		t.Errorf("expected error to list valid types (walking), got %q", err.Error())
	}
}

// TestCollectionsUpdateCmdTourType asserts --tour-type on update PATCHes the
// flat string field.
func TestCollectionsUpdateCmdTourType(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": map[string]interface{}{"id": 42, "tour_type": "cycling"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "update", "42", "--tour-type=cycling"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured["tour_type"] != "cycling" {
		t.Errorf("expected tour_type=\"cycling\" in PATCH body, got %v", captured["tour_type"])
	}
}

// TestProgressEnabledOffByDefault asserts progress is silent unless --progress
// is passed — dd(1)-style opt-in. Scripted callers (the common case) get clean
// stderr without needing any flag.
func TestProgressEnabledOffByDefault(t *testing.T) {
	orig := flagProgress
	t.Cleanup(func() { flagProgress = orig })

	flagProgress = false
	if progressEnabled() {
		t.Error("expected progressEnabled() to be false when --progress is not set")
	}

	flagProgress = true
	if !progressEnabled() {
		t.Error("expected progressEnabled() to be true when --progress is set")
	}
}
