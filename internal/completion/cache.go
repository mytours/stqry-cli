package completion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const ttl = time.Hour

// CacheEntry holds a resource ID and display name for tab-completion.
type CacheEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cacheFile struct {
	Items     []CacheEntry `json:"items"`
	FetchedAt time.Time    `json:"fetched_at,omitempty"`
}

// CachePath returns the path for a site+resource cache file. Exported for tests.
func CachePath(site, resource string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "stqry", "completions", site, resource+".json"), nil
}

// Load reads cached items. Returns (items, isStale, error).
// A missing file returns (nil, true, nil) — not an error.
func Load(site, resource string) ([]CacheEntry, bool, error) {
	path, err := CachePath(site, resource)
	if err != nil {
		return nil, true, nil
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, true, nil
	}
	if err != nil {
		return nil, true, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, true, err
	}

	var f cacheFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, true, err
	}

	stale := time.Since(info.ModTime()) > ttl
	return f.Items, stale, nil
}

// Save writes items to the cache file with the current timestamp.
func Save(site, resource string, items []CacheEntry) error {
	path, err := CachePath(site, resource)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f := cacheFile{Items: items, FetchedAt: time.Now()}
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
