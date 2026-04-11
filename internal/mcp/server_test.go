package mcp_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mytours/stqry-cli/internal/config"
	stqrymcp "github.com/mytours/stqry-cli/internal/mcp"
)

func TestConfigureProjectTool(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	err := stqrymcp.WriteProjectConfig("https://api.example.com", "tok123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "stqry.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("tok123")) {
		t.Error("expected token in stqry.yaml")
	}
	if !bytes.Contains(data, []byte("api.example.com")) {
		t.Error("expected api_url in stqry.yaml")
	}
}

func TestResolveClientFromDirConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "mytoken", APIURL: "https://api.example.com"}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	client, err := stqrymcp.ResolveClient("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Token != "mytoken" {
		t.Errorf("expected mytoken, got %s", client.Token)
	}
	_ = filepath.Join(dir, "stqry.yaml")
	_ = bytes.NewBuffer(nil)
}
