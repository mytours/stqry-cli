package cli

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillExportWritesZip(t *testing.T) {
	dir := t.TempDir()

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"skill", "export", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("skill export: %v", err)
	}

	zipPath := filepath.Join(dir, "stqry-skill.zip")
	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatalf("reading zip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opening zip: %v", err)
	}

	want := map[string]bool{
		"stqry-skill/SKILL.md":     false,
		"stqry-skill/SETUP.md":     false,
		"stqry-skill/REFERENCE.md": false,
		"stqry-skill/WORKFLOWS.md": false,
	}
	for _, f := range r.File {
		want[f.Name] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("zip missing: %s", name)
		}
	}
}

func TestSkillExportOutputContainsInstallInstructions(t *testing.T) {
	dir := t.TempDir()

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"skill", "export", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("skill export: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "stqry-skill.zip") {
		t.Errorf("expected output to mention zip path, got: %s", out)
	}
	if !strings.Contains(out, "Claude Desktop") {
		t.Errorf("expected Claude Desktop install instructions, got: %s", out)
	}
	if !strings.Contains(out, "Skill version:") {
		t.Errorf("expected Skill version in output, got: %s", out)
	}
}

func TestSkillExportDefaultsToCurrentDirectory(t *testing.T) {
	dir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("changing to temp dir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("restoring working directory: %v", err)
		}
	}()

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"skill", "export"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("skill export: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "stqry-skill.zip")); err != nil {
		t.Errorf("expected stqry-skill.zip in current directory: %v", err)
	}
}
