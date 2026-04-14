package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mytours/stqry-cli/internal/output"
)

func TestConfigValidateCmd(t *testing.T) {
	t.Run("all checks pass with valid token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("sites:\n  mysite:\n    token: goodtoken\n    api_url: "+srv.URL+"\n"), 0644)

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
		cmd.SetArgs([]string{"config", "validate"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("expected success, got error: %v\noutput:\n%s", execErr, out)
		}
		if !contains(out, "✓") {
			t.Errorf("expected pass symbols in output, got:\n%s", out)
		}
	})

	t.Run("fails when token is invalid", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("sites:\n  mysite:\n    token: badtoken\n    api_url: "+srv.URL+"\n"), 0644)

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
		cmd.SetArgs([]string{"config", "validate"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr == nil {
			t.Fatalf("expected error on invalid token, got nil\noutput:\n%s", out)
		}
		if !contains(out, "✗") {
			t.Errorf("expected fail symbol in output, got:\n%s", out)
		}
	})

	t.Run("--site flag targets named site directly", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		cfgDir := filepath.Join(tmpDir, ".config", "stqry")
		os.MkdirAll(cfgDir, 0755)
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("sites:\n  mysite:\n    token: goodtoken\n    api_url: "+srv.URL+"\n"), 0644)

		printer = &output.Printer{}
		r, w, _ := os.Pipe()
		origStdout := os.Stdout
		os.Stdout = w

		cmd := newRootCmd()
		cmd.SetArgs([]string{"--site", "mysite", "config", "validate"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr != nil {
			t.Fatalf("expected success, got error: %v\noutput:\n%s", execErr, out)
		}
		if !contains(out, "mysite") {
			t.Errorf("expected site name in output, got:\n%s", out)
		}
	})

	t.Run("fails when no site is configured", func(t *testing.T) {
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
		cmd.SetArgs([]string{"config", "validate"})
		execErr := cmd.Execute()

		w.Close()
		os.Stdout = origStdout
		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		r.Close()
		out := string(buf[:n])

		if execErr == nil {
			t.Fatalf("expected error with no site configured\noutput:\n%s", out)
		}
		if !contains(out, "✗") {
			t.Errorf("expected fail symbol in output, got:\n%s", out)
		}
	})
}
