package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mytours/stqry-cli/internal/completion"
	"github.com/spf13/cobra"
)

func TestCompletionBashCmd(t *testing.T) {
	setupTestHome(t, "http://unused")

	buf := &bytes.Buffer{}
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "bash"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("bash")) {
		t.Error("expected bash completion script in output")
	}
}

func TestCompletionZshCmd(t *testing.T) {
	setupTestHome(t, "http://unused")

	buf := &bytes.Buffer{}
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "zsh"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("zsh")) {
		t.Error("expected zsh completion script in output")
	}
}

func TestCompletionFishCmd(t *testing.T) {
	setupTestHome(t, "http://unused")
	buf := &bytes.Buffer{}
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "fish"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("fish")) {
		t.Error("expected fish completion script in output")
	}
}

func TestCompletionPowerShellCmd(t *testing.T) {
	setupTestHome(t, "http://unused")
	buf := &bytes.Buffer{}
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "powershell"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("powershell")) {
		t.Error("expected powershell completion script in output")
	}
}

func TestCompleteCollectionIDs_HitsCache(t *testing.T) {
	setupTestHome(t, "http://unused")

	items := []completion.CacheEntry{
		{ID: "42", Name: "city-tour"},
		{ID: "87", Name: "museum"},
	}
	if err := completion.Save("testsite", "collections", items); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	_ = cmd.ParseFlags([]string{"--site=testsite"})
	t.Cleanup(func() { flagSite = "" })
	collectionsCmd, _, _ := cmd.Find([]string{"collections"})
	getCmd, _, _ := collectionsCmd.Find([]string{"get"})

	results, directive := getCmd.ValidArgsFunction(getCmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive: %v", directive)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}
	if results[0] != "42\tcity-tour" {
		t.Errorf("unexpected first result: %q", results[0])
	}
}

func TestCompleteCollectionIDs_StaleHint(t *testing.T) {
	setupTestHome(t, "http://unused")

	items := []completion.CacheEntry{{ID: "1", Name: "old"}}
	if err := completion.Save("testsite", "collections", items); err != nil {
		t.Fatal(err)
	}
	path, _ := completion.CachePath("testsite", "collections")
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	os.Chtimes(path, twoHoursAgo, twoHoursAgo)

	cmd := newRootCmd()
	_ = cmd.ParseFlags([]string{"--site=testsite"})
	t.Cleanup(func() { flagSite = "" })
	collectionsCmd, _, _ := cmd.Find([]string{"collections"})
	getCmd, _, _ := collectionsCmd.Find([]string{"get"})

	results, _ := getCmd.ValidArgsFunction(getCmd, []string{}, "")
	last := results[len(results)-1]
	if last != "# cache stale — run: stqry completion refresh" {
		t.Errorf("expected stale hint, got: %q", last)
	}
}

func TestCompleteCollectionIDs_NoSite(t *testing.T) {
	flagSite = "" // ensure no leftover from prior tests
	t.Setenv("HOME", t.TempDir())

	cmd := newRootCmd()
	collectionsCmd, _, _ := cmd.Find([]string{"collections"})
	getCmd, _, _ := collectionsCmd.Find([]string{"get"})

	results, directive := getCmd.ValidArgsFunction(getCmd, []string{}, "")
	if len(results) != 0 {
		t.Errorf("expected no results, got: %v", results)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive: %v", directive)
	}
}

func TestCompleteScreenIDs_HitsCache(t *testing.T) {
	setupTestHome(t, "http://unused")
	flagSite = ""
	items := []completion.CacheEntry{{ID: "10", Name: "welcome"}}
	if err := completion.Save("testsite", "screens", items); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	_ = cmd.ParseFlags([]string{"--site=testsite"})
	t.Cleanup(func() { flagSite = "" })
	screensCmd, _, _ := cmd.Find([]string{"screens"})
	getCmd, _, _ := screensCmd.Find([]string{"get"})

	results, _ := getCmd.ValidArgsFunction(getCmd, []string{}, "")
	if len(results) != 1 || results[0] != "10\twelcome" {
		t.Errorf("unexpected results: %v", results)
	}
}

func TestCompleteMediaIDs_HitsCache(t *testing.T) {
	setupTestHome(t, "http://unused")
	flagSite = ""
	items := []completion.CacheEntry{{ID: "55", Name: "banner"}}
	if err := completion.Save("testsite", "media", items); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	_ = cmd.ParseFlags([]string{"--site=testsite"})
	t.Cleanup(func() { flagSite = "" })
	mediaCmd, _, _ := cmd.Find([]string{"media"})
	getCmd, _, _ := mediaCmd.Find([]string{"get"})

	results, _ := getCmd.ValidArgsFunction(getCmd, []string{}, "")
	if len(results) != 1 || results[0] != "55\tbanner" {
		t.Errorf("unexpected results: %v", results)
	}
}

func TestCompleteProjectIDs_HitsCache(t *testing.T) {
	setupTestHome(t, "http://unused")
	flagSite = ""
	items := []completion.CacheEntry{{ID: "1", Name: "main-project"}}
	if err := completion.Save("testsite", "projects", items); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	_ = cmd.ParseFlags([]string{"--site=testsite"})
	t.Cleanup(func() { flagSite = "" })
	projectsCmd, _, _ := cmd.Find([]string{"projects"})
	getCmd, _, _ := projectsCmd.Find([]string{"get"})

	results, _ := getCmd.ValidArgsFunction(getCmd, []string{}, "")
	if len(results) != 1 || results[0] != "1\tmain-project" {
		t.Errorf("unexpected results: %v", results)
	}
}

func TestCompletionStatusCmd(t *testing.T) {
	setupTestHome(t, "http://unused")
	flagSite = ""
	t.Cleanup(func() { flagSite = "" })

	completion.Save("testsite", "collections", []completion.CacheEntry{{ID: "1", Name: "a"}, {ID: "2", Name: "b"}})
	completion.Save("testsite", "screens", []completion.CacheEntry{{ID: "10", Name: "welcome"}})

	buf := &bytes.Buffer{}
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--site=testsite", "completion", "status"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "collections") {
		t.Errorf("expected 'collections' in output:\n%s", out)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("expected item count '2' in output:\n%s", out)
	}
	if !strings.Contains(out, "screens") {
		t.Errorf("expected 'screens' in output:\n%s", out)
	}
	if !strings.Contains(out, "not cached") {
		t.Errorf("expected 'not cached' for uncached resources:\n%s", out)
	}
}

func TestCompletionRefreshCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/public/collections":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"collections": []interface{}{
					map[string]interface{}{"id": float64(1), "name": "alpha"},
				},
				"meta": map[string]interface{}{"page": 1, "pages": 1, "per_page": 30, "count": 1},
			})
		case "/api/public/screens":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"screens": []interface{}{
					map[string]interface{}{"id": float64(2), "name": "welcome"},
				},
				"meta": map[string]interface{}{"page": 1, "pages": 1, "per_page": 30, "count": 1},
			})
		case "/api/public/media_items":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"media_items": []interface{}{
					map[string]interface{}{"id": float64(3), "name": "banner"},
				},
				"meta": map[string]interface{}{"page": 1, "pages": 1, "per_page": 30, "count": 1},
			})
		case "/api/public/projects":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"projects": []interface{}{
					map[string]interface{}{"id": float64(4), "name": "main"},
				},
				"meta": map[string]interface{}{"page": 1, "pages": 1, "per_page": 30, "count": 1},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "completion", "refresh"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	items, stale, err := completion.Load("testsite", "collections")
	if err != nil || stale || len(items) != 1 || items[0].ID != "1" {
		t.Errorf("collections cache wrong: items=%v stale=%v err=%v", items, stale, err)
	}
	items, stale, err = completion.Load("testsite", "screens")
	if err != nil || stale || len(items) != 1 || items[0].Name != "welcome" {
		t.Errorf("screens cache wrong: items=%v stale=%v err=%v", items, stale, err)
	}
}
