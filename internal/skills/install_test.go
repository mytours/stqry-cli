package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mytours/stqry-cli/internal/skills"
)

func TestInstallAll_CodeLayout(t *testing.T) {
	dir := t.TempDir()
	if err := skills.InstallAll(dir, "v1.0.0"); err != nil {
		t.Fatalf("InstallAll: %v", err)
	}

	// Each skill is a flat .md file with frontmatter.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected files to be installed")
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("reading file %s: %v", e.Name(), err)
		}
		if !strings.HasPrefix(string(data), "---\n") {
			t.Errorf("file %s missing frontmatter", e.Name())
		}
		if !strings.Contains(string(data), "skill_version: v1.0.0") {
			t.Errorf("file %s missing skill_version", e.Name())
		}
	}
}

func TestInstallAll_Overwrites(t *testing.T) {
	dir := t.TempDir()

	// Install once.
	skills.InstallAll(dir, "v1.0.0")

	// Overwrite with a different version.
	if err := skills.InstallAll(dir, "v2.0.0"); err != nil {
		t.Fatalf("second InstallAll: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("reading file %s: %v", e.Name(), err)
		}
		if !strings.Contains(string(data), "skill_version: v2.0.0") {
			t.Errorf("expected v2.0.0 after overwrite in %s", e.Name())
		}
	}
}

