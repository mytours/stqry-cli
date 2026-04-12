package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mytours/stqry-cli/internal/config"
)

func TestCheckGlobalConfig(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		tmp := t.TempDir()
		cfgPath := filepath.Join(tmp, "config.yaml")
		os.WriteFile(cfgPath, []byte("sites: {}\n"), 0644)

		r := checkGlobalConfig(cfgPath)
		if r.status != statusPass {
			t.Errorf("expected pass, got %s: %s", r.status, r.message)
		}
		if r.detail == "" {
			t.Error("expected detail to contain path")
		}
	})

	t.Run("file missing", func(t *testing.T) {
		r := checkGlobalConfig("/does/not/exist/config.yaml")
		if r.status != statusFail {
			t.Errorf("expected fail, got %s", r.status)
		}
	})
}

func TestCheckDirectoryConfig(t *testing.T) {
	t.Run("stqry.yaml found", func(t *testing.T) {
		tmp := t.TempDir()
		os.WriteFile(filepath.Join(tmp, "stqry.yaml"), []byte("site: testsite\n"), 0644)

		r := checkDirectoryConfig(tmp)
		if r.status != statusPass {
			t.Errorf("expected pass, got %s: %s", r.status, r.message)
		}
	})

	t.Run("not found", func(t *testing.T) {
		tmp := t.TempDir()
		r := checkDirectoryConfig(tmp)
		if r.status != statusFail {
			t.Errorf("expected fail, got %s", r.status)
		}
	})
}

func TestCheckSiteResolved(t *testing.T) {
	t.Run("resolves from global config", func(t *testing.T) {
		global := &config.GlobalConfig{
			Sites: map[string]*config.Site{
				"prod": {Token: "tok", APIURL: "https://api-us.stqry.com"},
			},
		}
		r := checkSiteResolved(global, "prod", nil)
		if r.status != statusPass {
			t.Errorf("expected pass, got %s: %s", r.status, r.message)
		}
	})

	t.Run("no site available", func(t *testing.T) {
		global := &config.GlobalConfig{Sites: map[string]*config.Site{}}
		r := checkSiteResolved(global, "", nil)
		if r.status != statusFail {
			t.Errorf("expected fail, got %s", r.status)
		}
	})
}

func TestCheckStatusSymbols(t *testing.T) {
	tests := []struct {
		status checkStatus
		want   string
	}{
		{statusPass, "✓"},
		{statusFail, "✗"},
		{statusSkip, "-"},
		{statusInfo, "ℹ"},
		{statusWarn, "⚠"},
		{checkStatus("bogus"), "?"},
	}
	for _, tt := range tests {
		if got := doctorSymbol(tt.status); got != tt.want {
			t.Errorf("doctorSymbol(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}
