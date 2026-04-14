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

func TestFindDirectoryConfigFindsStqryYaml(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "stqry.yaml"), []byte("site: bobs\n"), 0644)

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

func TestFindDirectoryConfigFindsStqryYml(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "stqry.yml"), []byte("site: ymlsite\n"), 0644)

	dirCfg, err := FindDirectoryConfig(root)
	if err != nil {
		t.Fatalf("FindDirectoryConfig: %v", err)
	}
	if dirCfg.Site != "ymlsite" {
		t.Errorf("expected site ymlsite, got %s", dirCfg.Site)
	}
}

func TestFindDirectoryConfigPrefersYamlOverYml(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "stqry.yaml"), []byte("site: yamlwins\n"), 0644)
	os.WriteFile(filepath.Join(root, "stqry.yml"), []byte("site: ymllosses\n"), 0644)

	dirCfg, err := FindDirectoryConfig(root)
	if err != nil {
		t.Fatalf("FindDirectoryConfig: %v", err)
	}
	if dirCfg.Site != "yamlwins" {
		t.Errorf("expected site yamlwins, got %s", dirCfg.Site)
	}
}

func TestSaveDirectoryConfigWritesStqryYaml(t *testing.T) {
	dir := t.TempDir()
	cfg := &DirectoryConfig{Site: "testsite"}

	if err := SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatalf("SaveDirectoryConfig: %v", err)
	}

	expectedPath := filepath.Join(dir, "stqry.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", expectedPath)
	}

	// Hidden dir should NOT be created.
	hiddenPath := filepath.Join(dir, ".stqry", "config.yaml")
	if _, err := os.Stat(hiddenPath); err == nil {
		t.Errorf("expected .stqry/config.yaml NOT to exist, but it does")
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

	// --site flag resolves from global config.
	site, err := ResolveSite(global, "bobs", &DirectoryConfig{})
	if err != nil {
		t.Fatalf("ResolveSite with flag: %v", err)
	}
	if site.Token != "ABC123" {
		t.Errorf("expected ABC123, got %s", site.Token)
	}

	// Directory config referencing a named site.
	site, err = ResolveSite(global, "", &DirectoryConfig{Site: "bobs"})
	if err != nil {
		t.Fatalf("ResolveSite with dir site name: %v", err)
	}
	if site.Token != "ABC123" {
		t.Errorf("expected ABC123, got %s", site.Token)
	}

	// Directory config with inline credentials.
	site, err = ResolveSite(global, "", &DirectoryConfig{Token: "INLINE", APIURL: "https://api-eu.stqry.com"})
	if err != nil {
		t.Fatalf("ResolveSite with inline dir config: %v", err)
	}
	if site.Token != "INLINE" {
		t.Errorf("expected INLINE, got %s", site.Token)
	}

	// No site specified.
	_, err = ResolveSite(global, "", &DirectoryConfig{})
	if err == nil {
		t.Fatal("expected error when no site specified")
	}

	// Unknown site in flag.
	_, err = ResolveSite(global, "unknown", &DirectoryConfig{})
	if err == nil {
		t.Fatal("expected error for unknown site")
	}
}

func TestResolveSiteEnvVar(t *testing.T) {
	t.Setenv("STQRY_SITE", "env-site")
	global := &GlobalConfig{
		Sites: map[string]*Site{
			"env-site": {Token: "tok", APIURL: "https://api.example.com"},
		},
	}
	site, err := ResolveSite(global, "", &DirectoryConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if site.Token != "tok" {
		t.Errorf("expected tok, got %s", site.Token)
	}
}

func TestFindDirectoryConfigWithPath(t *testing.T) {
	t.Run("finds stqry.yaml and returns path", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, "stqry.yaml"), []byte("site: mysite\n"), 0644); err != nil {
			t.Fatal(err)
		}
		cfg, path, err := FindDirectoryConfigWithPath(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Site != "mysite" {
			t.Errorf("expected site %q, got %q", "mysite", cfg.Site)
		}
		if path != filepath.Join(tmp, "stqry.yaml") {
			t.Errorf("expected path %q, got %q", filepath.Join(tmp, "stqry.yaml"), path)
		}
	})

	t.Run("returns empty path when not found", func(t *testing.T) {
		tmp := t.TempDir()
		_, path, err := FindDirectoryConfigWithPath(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if path != "" {
			t.Errorf("expected empty path, got %q", path)
		}
	})
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
