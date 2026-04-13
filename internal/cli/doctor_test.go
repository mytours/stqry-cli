package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mytours/stqry-cli/internal/config"
	"github.com/mytours/stqry-cli/internal/skills"
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
		r, site := checkSiteResolved(global, "prod", nil)
		if r.status != statusPass {
			t.Errorf("expected pass, got %s: %s", r.status, r.message)
		}
		if site == nil || site.APIURL != "https://api-us.stqry.com" {
			t.Errorf("expected resolved site with correct APIURL, got %v", site)
		}
	})

	t.Run("no site available", func(t *testing.T) {
		global := &config.GlobalConfig{Sites: map[string]*config.Site{}}
		r, site := checkSiteResolved(global, "", nil)
		if r.status != statusFail {
			t.Errorf("expected fail, got %s", r.status)
		}
		if site != nil {
			t.Errorf("expected nil site on failure, got %v", site)
		}
	})
}

func TestCheckAPIReachable(t *testing.T) {
	t.Run("server responds", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		r := checkAPIReachable(srv.URL, srv.Client())
		if r.status != statusPass {
			t.Errorf("expected pass, got %s: %s", r.status, r.message)
		}
	})

	t.Run("server unreachable", func(t *testing.T) {
		r := checkAPIReachable("http://127.0.0.1:1", http.DefaultClient)
		if r.status != statusFail {
			t.Errorf("expected fail, got %s", r.status)
		}
	})

	t.Run("malformed URL", func(t *testing.T) {
		r := checkAPIReachable("not-a-url", http.DefaultClient)
		if r.status != statusFail {
			t.Errorf("expected fail, got %s", r.status)
		}
	})
}

func TestCheckTokenValid(t *testing.T) {
	t.Run("valid token returns 200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		r := checkTokenValid(srv.URL, "good-token", srv.Client())
		if r.status != statusPass {
			t.Errorf("expected pass, got %s: %s", r.status, r.message)
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Api-Token") == "" {
				t.Error("expected X-Api-Token header to be set")
			}
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		r := checkTokenValid(srv.URL, "bad-token", srv.Client())
		if r.status != statusFail {
			t.Errorf("expected fail, got %s: %s", r.status, r.message)
		}
	})

	t.Run("forbidden token returns 403", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		r := checkTokenValid(srv.URL, "low-priv-token", srv.Client())
		if r.status != statusFail {
			t.Errorf("expected fail, got %s: %s", r.status, r.message)
		}
	})
}

func TestCheckRegion(t *testing.T) {
	tests := []struct {
		apiURL     string
		wantMsg    string
		wantStatus checkStatus
	}{
		{"https://api-us.stqry.com", "us", statusInfo},
		{"https://api-eu.stqry.com", "eu", statusInfo},
		{"https://custom.example.com", "custom.example.com", statusInfo},
	}
	for _, tt := range tests {
		r := checkRegion(tt.apiURL)
		if r.status != tt.wantStatus {
			t.Errorf("checkRegion(%q): expected %s, got %s", tt.apiURL, tt.wantStatus, r.status)
		}
		if !contains(r.message, tt.wantMsg) {
			t.Errorf("checkRegion(%q): expected message to contain %q, got %q", tt.apiURL, tt.wantMsg, r.message)
		}
	}
}

func TestCheckCLIVersion(t *testing.T) {
	t.Run("dev build skips check", func(t *testing.T) {
		r := checkCLIVersion("dev", "", nil)
		if r.status != statusInfo {
			t.Errorf("expected info for dev build, got %s", r.status)
		}
		if !contains(r.message, "development") {
			t.Errorf("expected message to mention development build, got: %s", r.message)
		}
	})

	t.Run("up to date", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"tag_name": "v0.3.1"})
		}))
		defer srv.Close()

		r := checkCLIVersion("v0.3.1", srv.URL, srv.Client())
		if r.status != statusPass {
			t.Errorf("expected pass, got %s: %s", r.status, r.message)
		}
	})

	t.Run("update available", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"tag_name": "v0.3.2"})
		}))
		defer srv.Close()

		r := checkCLIVersion("v0.3.1", srv.URL, srv.Client())
		if r.status != statusWarn {
			t.Errorf("expected warn (update available), got %s", r.status)
		}
		if !contains(r.detail, "v0.3.2") {
			t.Errorf("expected detail to contain latest version, got: %s", r.detail)
		}
	})

	t.Run("github unreachable", func(t *testing.T) {
		r := checkCLIVersion("v0.3.1", "http://127.0.0.1:1", http.DefaultClient)
		if r.status != statusWarn {
			t.Errorf("expected warn for unreachable GitHub, got %s", r.status)
		}
	})

	t.Run("malformed response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"tag_name": ""})
		}))
		defer srv.Close()

		r := checkCLIVersion("v0.3.1", srv.URL, srv.Client())
		if r.status != statusWarn {
			t.Errorf("expected warn for empty tag, got %s", r.status)
		}
	})

	t.Run("invalid json response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer srv.Close()

		r := checkCLIVersion("v0.3.1", srv.URL, srv.Client())
		if r.status != statusWarn {
			t.Errorf("expected warn for invalid JSON, got %s", r.status)
		}
		if r.detail == "" {
			t.Error("expected detail to contain decode error")
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

func TestPrintDoctorResults(t *testing.T) {
	results := []checkResult{
		{group: "Config", name: "Global config exists", status: statusPass, message: "~/.config/stqry/config.yaml"},
		{group: "Config", name: "Directory config found", status: statusFail, message: ""},
		{group: "API", name: "API reachable", status: statusPass, message: "api-us.stqry.com"},
		{group: "API", name: "Region", status: statusInfo, message: "us"},
		{group: "Version", name: "CLI version", status: statusSkip, message: ""},
		{group: "Version", name: "CLI update", status: statusWarn, message: "v0.3.2 available"},
	}

	var buf bytes.Buffer
	printDoctorResults(&buf, results, false)
	out := buf.String()

	// Group headers appear
	if !contains(out, "Config\n") {
		t.Errorf("expected Config group header, got:\n%s", out)
	}
	if !contains(out, "API\n") {
		t.Errorf("expected API group header, got:\n%s", out)
	}
	if !contains(out, "Version\n") {
		t.Errorf("expected Version group header, got:\n%s", out)
	}
	// Config header appears exactly once
	if strings.Count(out, "Config\n") != 1 {
		t.Errorf("expected Config header exactly once, got:\n%s", out)
	}
	// All symbols appear
	if !contains(out, "✓") {
		t.Errorf("expected pass symbol, got:\n%s", out)
	}
	if !contains(out, "✗") {
		t.Errorf("expected fail symbol, got:\n%s", out)
	}
	if !contains(out, "-") {
		t.Errorf("expected skip symbol, got:\n%s", out)
	}
	if !contains(out, "⚠") {
		t.Errorf("expected warn symbol, got:\n%s", out)
	}
	// Message content appears in parens
	if !contains(out, "~/.config/stqry/config.yaml") {
		t.Errorf("expected path message in output, got:\n%s", out)
	}
	if !contains(out, "api-us.stqry.com") {
		t.Errorf("expected host in output, got:\n%s", out)
	}
	// Check name appears
	if !contains(out, "Global config exists") {
		t.Errorf("expected check name in output, got:\n%s", out)
	}
	// Empty message produces no parens
	if contains(out, "Directory config found ()") {
		t.Errorf("empty message should not produce empty parens, got:\n%s", out)
	}
}

func TestPrintDoctorResultsVerbose(t *testing.T) {
	results := []checkResult{
		{
			group:   "Config",
			name:    "Global config exists",
			status:  statusPass,
			message: "~/.config/stqry/config.yaml",
			detail:  "Path: /Users/glen/.config/stqry/config.yaml",
		},
	}

	var buf bytes.Buffer
	printDoctorResults(&buf, results, true)
	out := buf.String()

	if !contains(out, "Path: /Users/glen") {
		t.Errorf("expected detail line in verbose output, got:\n%s", out)
	}
}

func TestDoctorCmd_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("STQRY_CONFIG_HOME", filepath.Join(tmpDir, ".config", "stqry"))

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"doctor"})
	_ = cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes := make([]byte, 8192)
	n, _ := r.Read(outBytes)
	r.Close()
	out := string(outBytes[:n])

	if !contains(out, "Config") {
		t.Errorf("expected Config section in output, got:\n%s", out)
	}
	if !contains(out, "✗") {
		t.Errorf("expected at least one fail symbol, got:\n%s", out)
	}
	if !contains(out, "Global config exists") {
		t.Errorf("expected global config check in output, got:\n%s", out)
	}
}

func TestDoctorCmd_APISkipped_WhenNoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("STQRY_CONFIG_HOME", filepath.Join(tmpDir, ".config", "stqry"))

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"doctor"})
	_ = cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes := make([]byte, 8192)
	n, _ := r.Read(outBytes)
	r.Close()
	out := string(outBytes[:n])

	// API checks should be skipped (shown with "-") when no site is configured
	if !contains(out, "API") {
		t.Errorf("expected API section in output, got:\n%s", out)
	}
	if !contains(out, "-") {
		t.Errorf("expected skip symbol for API checks, got:\n%s", out)
	}
}

func TestCheckOneInstalledSkill_PassMessageIsJustUpToDate(t *testing.T) {
	dir := t.TempDir()
	if err := skills.InstallAll(dir, skills.LayoutCode, "v1.0.0"); err != nil {
		t.Fatalf("InstallAll: %v", err)
	}

	skillNames, err := skills.EmbeddedSkillNames()
	if err != nil || len(skillNames) == 0 {
		t.Fatalf("EmbeddedSkillNames: %v", err)
	}

	r := checkOneInstalledSkill(dir, "test", skillNames[0])
	if r.status != statusPass {
		t.Fatalf("expected pass, got %s: %s", r.status, r.message)
	}
	if r.message != "up to date" {
		t.Errorf("pass message should be 'up to date', got %q (skill name must not be repeated)", r.message)
	}
}

func TestDoctorShowsSkillsGroup(t *testing.T) {
	// Install a skill with a bad hash so the Skills group appears with a warn.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	setupTestHome(t, srv.URL)

	// Write a stale skill into the global Claude commands dir (under the HOME
	// that setupTestHome just configured via t.Setenv).
	home, _ := os.UserHomeDir()
	commandsDir := filepath.Join(home, ".claude", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatalf("creating commands dir: %v", err)
	}
	staleContent := "---\nskill_version: v0.1.0\nskill_hash: 000000000000\ngenerated_by: stqry-cli\n---\n# stale\n"
	if err := os.WriteFile(filepath.Join(commandsDir, "stqry-reference.md"), []byte(staleContent), 0644); err != nil {
		t.Fatalf("writing stale skill: %v", err)
	}

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"doctor"})
	cmd.Execute() // ignore exit code — doctor exits non-zero on warn

	w.Close()
	os.Stdout = origStdout

	outBytes := make([]byte, 16384)
	n, _ := r.Read(outBytes)
	r.Close()
	out := string(outBytes[:n])

	// Check for the "Skills" group header — printDoctorResults writes group names
	// as bare lines ("Skills\n"), so we look for that exact pattern.
	if !strings.Contains(out, "\nSkills\n") {
		t.Errorf("expected 'Skills' group header in doctor output, got:\n%s", out)
	}
}
