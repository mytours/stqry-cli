package completion_test

import (
	"os"
	"testing"
	"time"

	"github.com/mytours/stqry-cli/internal/completion"
)

func TestSaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

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
	t.Setenv("HOME", t.TempDir())

	items, stale, err := completion.Load("nosite", "collections")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !stale {
		t.Error("expected stale for missing file")
	}
	if len(items) != 0 {
		t.Errorf("expected no items, got %v", items)
	}
}

func TestIsStale(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

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
