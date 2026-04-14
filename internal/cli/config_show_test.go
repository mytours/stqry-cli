package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mytours/stqry-cli/internal/output"
)

func TestConfigShowCmd(t *testing.T) {
	t.Run("shows global config and no directory config", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`sites:
  mysite:
    token: abc12345token
    api_url: https://api-us.stqry.com
`), 0644)

		printer = &output.Printer{}
		r, w, _ := os.Pipe()
		origStdout := os.Stdout
		os.Stdout = w

		cmd := newRootCmd()
		cmd.SetArgs([]string{"config", "show"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("Execute: %v", execErr)
		}
		if !contains(out, "mysite") {
			t.Errorf("expected output to contain site name, got:\n%s", out)
		}
		if !contains(out, "https://api-us.stqry.com") {
			t.Errorf("expected output to contain API URL, got:\n%s", out)
		}
		if !contains(out, "abc12345") {
			t.Errorf("expected output to contain masked token prefix, got:\n%s", out)
		}
		if !contains(out, "(none)") {
			t.Errorf("expected output to contain '(none)' for missing directory config, got:\n%s", out)
		}
	})

	t.Run("shows directory config source when stqry.yaml present", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`sites:
  mysite:
    token: tok123456
    api_url: https://api-us.stqry.com
`), 0644)

		workDir := t.TempDir()
		os.WriteFile(filepath.Join(workDir, "stqry.yaml"), []byte("site: mysite\n"), 0644)
		origWD, _ := os.Getwd()
		os.Chdir(workDir)
		defer os.Chdir(origWD)

		printer = &output.Printer{}
		r, w, _ := os.Pipe()
		origStdout := os.Stdout
		os.Stdout = w

		cmd := newRootCmd()
		cmd.SetArgs([]string{"config", "show"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("Execute: %v", execErr)
		}
		if !contains(out, "stqry.yaml") {
			t.Errorf("expected output to contain source label, got:\n%s", out)
		}
	})

	t.Run("json mode outputs structured data", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`sites:
  mysite:
    token: abc12345token
    api_url: https://api-us.stqry.com
`), 0644)

		printer = &output.Printer{JSON: true}
		r, w, _ := os.Pipe()
		origStdout := os.Stdout
		os.Stdout = w

		cmd := newRootCmd()
		cmd.SetArgs([]string{"--json", "config", "show"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("Execute: %v", execErr)
		}
		var envelope map[string]interface{}
		if err := json.Unmarshal(buf[:n], &envelope); err != nil {
			t.Fatalf("expected valid JSON, got:\n%s\nerror: %v", out, err)
		}
		if _, ok := envelope["data"]; !ok {
			t.Errorf("expected JSON envelope with 'data' key, got: %s", out)
		}
	})

	t.Run("shows parse error when stqry.yaml is invalid", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`sites:
  mysite:
    token: abc12345token
    api_url: https://api-us.stqry.com
`), 0644)

		workDir := t.TempDir()
		os.WriteFile(filepath.Join(workDir, "stqry.yaml"), []byte(":\ninvalid: yaml: content:\n"), 0644)
		origWD, _ := os.Getwd()
		os.Chdir(workDir)
		defer os.Chdir(origWD)

		printer = &output.Printer{}
		r, w, _ := os.Pipe()
		origStdout := os.Stdout
		os.Stdout = w

		cmd := newRootCmd()
		cmd.SetArgs([]string{"config", "show"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("Execute: %v", execErr)
		}
		if !contains(out, "parse error") {
			t.Errorf("expected output to contain 'parse error', got:\n%s", out)
		}
	})

	t.Run("shows not configured when no site is active", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("sites: {}\n"), 0644)

		printer = &output.Printer{}
		r, w, _ := os.Pipe()
		origStdout := os.Stdout
		os.Stdout = w

		cmd := newRootCmd()
		cmd.SetArgs([]string{"config", "show"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("Execute: %v", execErr)
		}
		if !contains(out, "(not configured)") {
			t.Errorf("expected '(not configured)', got:\n%s", out)
		}
	})

	t.Run("json mode outputs null active_site when no site configured", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("sites: {}\n"), 0644)

		printer = &output.Printer{JSON: true}
		r, w, _ := os.Pipe()
		origStdout := os.Stdout
		os.Stdout = w

		cmd := newRootCmd()
		cmd.SetArgs([]string{"--json", "config", "show"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("Execute: %v", execErr)
		}
		var envelope map[string]interface{}
		if err := json.Unmarshal(buf[:n], &envelope); err != nil {
			t.Fatalf("expected valid JSON, got:\n%s\nerror: %v", out, err)
		}
		data, ok := envelope["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected JSON envelope with 'data' map, got: %s", out)
		}
		if activeSite, exists := data["active_site"]; exists && activeSite != nil {
			t.Errorf("expected active_site to be null, got: %v", activeSite)
		}
	})
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc12345token", "abc12345..."},
		{"short", "short"},
		{"exactly8", "exactly8"},
		{"exactly8!", "exactly8..."},
	}
	for _, tt := range tests {
		got := maskToken(tt.input)
		if got != tt.want {
			t.Errorf("maskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRegionFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://api-us.stqry.com", "us"},
		{"https://api-eu.stqry.com", "eu"},
		{"https://api-sg.stqry.com", "sg"},
		{"https://staging.example.com", ""},
	}
	for _, tt := range tests {
		got := regionFromURL(tt.url)
		if got != tt.want {
			t.Errorf("regionFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
