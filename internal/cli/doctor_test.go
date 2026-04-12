package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		if r.status != statusFail {
			t.Errorf("expected fail (update available), got %s", r.status)
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
