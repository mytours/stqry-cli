package completion_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mytours/stqry-cli/internal/completion"
)

func setCacheTempDir(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", "") // force HOME-relative path on Linux
}

func TestSaveAndLoad(t *testing.T) {
	setCacheTempDir(t)

	items := []completion.CacheEntry{
		{ID: "42", Name: "city-tour"},
		{ID: "87", Name: "museum"},
	}
	if err := completion.Save("mysite", "collections", items); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, stale, err := completion.Load("mysite", "collections")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if stale {
		t.Error("expected fresh, got stale")
	}
	if len(got) != 2 || got[0].ID != "42" || got[1].Name != "museum" {
		t.Errorf("unexpected items: %v", got)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	setCacheTempDir(t)

	items, stale, err := completion.Load("nosite", "collections")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !stale {
		t.Error("expected stale for missing file")
	}
	if items != nil {
		t.Errorf("expected nil items for missing file, got %v", items)
	}
}

func TestLoad_CorruptFile(t *testing.T) {
	setCacheTempDir(t)

	path, err := completion.CachePath("mysite", "collections")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not valid json{{{{"), 0644); err != nil {
		t.Fatal(err)
	}

	items, stale, err := completion.Load("mysite", "collections")
	if err == nil {
		t.Error("expected error for corrupt cache file, got nil")
	}
	if items != nil {
		t.Errorf("expected nil items for corrupt file, got %v", items)
	}
	_ = stale // stale value is unspecified for error case
}

func TestIsStale(t *testing.T) {
	setCacheTempDir(t)

	items := []completion.CacheEntry{{ID: "1", Name: "a"}}

	if err := completion.Save("mysite", "collections", items); err != nil {
		t.Fatal(err)
	}

	path, _ := completion.CachePath("mysite", "collections")
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, twoHoursAgo, twoHoursAgo); err != nil {
		t.Fatal(err)
	}

	_, stale, err := completion.Load("mysite", "collections")
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Error("expected stale after 2 hours")
	}
}
