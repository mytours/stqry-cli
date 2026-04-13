package cli

import (
	"bytes"
	"os"
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
