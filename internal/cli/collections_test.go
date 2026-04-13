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
