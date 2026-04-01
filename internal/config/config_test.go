package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGlobalConfig(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(globalPath, []byte(`sites:
  bobs:
    token: ABC123
    api_url: https://api-us.area360.com
  museum:
    token: DEF456
    api_url: https://api-ca.area360.com
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadGlobalConfig(globalPath)
	if err != nil {
		t.Fatalf("LoadGlobalConfig: %v", err)
	}
	if len(cfg.Sites) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(cfg.Sites))
	}
	if cfg.Sites["bobs"].Token != "ABC123" {
		t.Errorf("expected token ABC123, got %s", cfg.Sites["bobs"].Token)
	}
	if cfg.Sites["bobs"].APIURL != "https://api-us.area360.com" {
		t.Errorf("expected api_url https://api-us.area360.com, got %s", cfg.Sites["bobs"].APIURL)
	}
}

func TestLoadGlobalConfigMissing(t *testing.T) {
	cfg, err := LoadGlobalConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("should not error on missing file: %v", err)
	}
	if len(cfg.Sites) != 0 {
		t.Fatalf("expected 0 sites, got %d", len(cfg.Sites))
	}
}

func TestFindDirectoryConfig(t *testing.T) {
	root := t.TempDir()
	stqryDir := filepath.Join(root, ".stqry")
	os.MkdirAll(stqryDir, 0755)
	os.WriteFile(filepath.Join(stqryDir, "config.yaml"), []byte("site: bobs\n"), 0644)

	deepDir := filepath.Join(root, "sub", "deep")
	os.MkdirAll(deepDir, 0755)

	dirCfg, err := FindDirectoryConfig(deepDir)
	if err != nil {
		t.Fatalf("FindDirectoryConfig: %v", err)
	}
	if dirCfg.Site != "bobs" {
		t.Errorf("expected site bobs, got %s", dirCfg.Site)
	}
}

func TestFindDirectoryConfigNotFound(t *testing.T) {
	dir := t.TempDir()
	dirCfg, err := FindDirectoryConfig(dir)
	if err != nil {
		t.Fatalf("should not error: %v", err)
	}
	if dirCfg.Site != "" {
		t.Errorf("expected empty site, got %s", dirCfg.Site)
	}
}

func TestResolveSite(t *testing.T) {
	global := &GlobalConfig{
		Sites: map[string]*Site{
			"bobs": {Token: "ABC123", APIURL: "https://api-us.area360.com"},
		},
	}

	site, err := ResolveSite(global, "bobs", "")
	if err != nil {
		t.Fatalf("ResolveSite with flag: %v", err)
	}
	if site.Token != "ABC123" {
		t.Errorf("expected ABC123, got %s", site.Token)
	}

	site, err = ResolveSite(global, "", "bobs")
	if err != nil {
		t.Fatalf("ResolveSite with dir: %v", err)
	}
	if site.Token != "ABC123" {
		t.Errorf("expected ABC123, got %s", site.Token)
	}

	_, err = ResolveSite(global, "", "")
	if err == nil {
		t.Fatal("expected error when no site specified")
	}

	_, err = ResolveSite(global, "unknown", "")
	if err == nil {
		t.Fatal("expected error for unknown site")
	}
}

func TestSaveGlobalConfig(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "config.yaml")

	cfg := &GlobalConfig{
		Sites: map[string]*Site{
			"test": {Token: "TOK", APIURL: "https://example.com"},
		},
	}
	err := SaveGlobalConfig(cfg, globalPath)
	if err != nil {
		t.Fatalf("SaveGlobalConfig: %v", err)
	}

	loaded, err := LoadGlobalConfig(globalPath)
	if err != nil {
		t.Fatalf("LoadGlobalConfig after save: %v", err)
	}
	if loaded.Sites["test"].Token != "TOK" {
		t.Errorf("expected TOK, got %s", loaded.Sites["test"].Token)
	}
}
