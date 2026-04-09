package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mytours/stqry-cli/internal/config"
)

func TestResolveAPIURL(t *testing.T) {
	tests := []struct {
		name      string
		region    string
		apiURL    string
		wantURL   string
		wantError bool
		errorMsg  string
	}{
		{
			name:    "region us",
			region:  "us",
			wantURL: "https://api-us.stqry.com",
		},
		{
			name:    "region eu",
			region:  "eu",
			wantURL: "https://api-eu.stqry.com",
		},
		{
			name:    "custom apiURL takes precedence over region",
			region:  "us",
			apiURL:  "https://staging.example.com",
			wantURL: "https://staging.example.com",
		},
		{
			name:      "no region and no apiURL returns error",
			wantError: true,
			errorMsg:  "either --region or --api-url is required",
		},
		{
			name:      "unknown region returns error",
			region:    "xx",
			wantError: true,
			errorMsg:  `unknown region "xx"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAPIURL(tt.region, tt.apiURL)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errorMsg != "" {
					if !contains(err.Error(), tt.errorMsg) {
						t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantURL {
				t.Errorf("expected URL %q, got %q", tt.wantURL, got)
			}
		})
	}
}

func TestConfigAddSiteCmd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Pre-set globalConfig so the command can find it before PersistentPreRunE
	// overwrites it; the pre-run will load from the (empty) temp HOME, which
	// yields an empty config — that is fine because add-site only needs Sites
	// to be non-nil, which LoadGlobalConfig guarantees.
	globalConfig = &config.GlobalConfig{Sites: make(map[string]*config.Site)}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config", "add-site", "--name=test", "--token=tok123", "--region=us"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Load the saved config from the temp dir and verify the site was written.
	cfgPath := filepath.Join(tmpDir, ".config", "stqry", "config.yaml")
	savedCfg, err := config.LoadGlobalConfig(cfgPath)
	if err != nil {
		t.Fatalf("loading saved config: %v", err)
	}

	site, ok := savedCfg.Sites["test"]
	if !ok {
		t.Fatalf("site %q not found in saved config; sites = %v", "test", savedCfg.Sites)
	}
	if site.Token != "tok123" {
		t.Errorf("expected token %q, got %q", "tok123", site.Token)
	}
	if site.APIURL != "https://api-us.stqry.com" {
		t.Errorf("expected APIURL %q, got %q", "https://api-us.stqry.com", site.APIURL)
	}
}

func TestConfigListSitesCmd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Write a config with two sites to the temp HOME so PersistentPreRunE
	// loads it and populates globalConfig before list-sites runs.
	cfgDir := filepath.Join(tmpDir, ".config", "stqry")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}
	cfgYAML := `sites:
  alpha:
    token: tok-alpha
    api_url: https://api-us.stqry.com
  beta:
    token: tok-beta
    api_url: https://api-eu.stqry.com
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgYAML), 0644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	// Capture stdout via os.Pipe because Printer writes directly to os.Stdout.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"config", "list-sites"})
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
	if !contains(out, "beta") {
		t.Errorf("expected output to contain %q, got:\n%s", "beta", out)
	}
}

// contains is a simple substring helper to avoid importing strings in test
// helper calls scattered across tests.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
