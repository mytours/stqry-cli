package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupClaudeInstallsWithFrontmatter(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := t.TempDir()

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"setup", "claude"})

	origDir, _ := os.Getwd()
	os.Chdir(cwd)
	defer os.Chdir(origDir)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("setup claude: %v", err)
	}

	commandsDir := filepath.Join(cwd, ".claude", "commands")
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		t.Fatalf("reading commands dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no skills installed")
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(commandsDir, e.Name()))
		if err != nil {
			t.Fatalf("reading installed file %s: %v", e.Name(), err)
		}
		if !strings.HasPrefix(string(data), "---\n") {
			t.Errorf("%s: missing frontmatter", e.Name())
		}
		if !strings.Contains(string(data), "skill_hash:") {
			t.Errorf("%s: missing skill_hash", e.Name())
		}
	}
}

func TestSetupClaudeGlobalFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"setup", "claude", "--global"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("setup claude --global: %v", err)
	}

	globalDir := filepath.Join(home, ".claude", "commands")
	entries, err := os.ReadDir(globalDir)
	if err != nil {
		t.Fatalf("reading global commands dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no skills installed globally")
	}
}

func TestSetupClaudeDesktopExportsToDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	exportDir := t.TempDir()

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"setup", "claude", "--desktop", "--dir", exportDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("setup claude --desktop --dir: %v", err)
	}

	// Flat .md files should be written to exportDir (not subfolder/SKILL.md).
	entries, err := os.ReadDir(exportDir)
	if err != nil {
		t.Fatalf("reading export dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no skill files exported")
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			t.Errorf("expected .md file, got %s", e.Name())
		}
		// Must not create subfolders.
		if e.IsDir() {
			t.Errorf("expected flat .md file, got directory %s", e.Name())
		}
	}

	// Output must include UI instructions.
	out := buf.String()
	if !strings.Contains(out, "Claude Desktop") {
		t.Error("expected 'Claude Desktop' in output")
	}
	if !strings.Contains(out, "Settings") {
		t.Error("expected 'Settings' mention in UI instructions")
	}
}

func TestSetupClaudeDesktopDefaultsToDownloads(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"setup", "claude", "--desktop"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("setup claude --desktop: %v", err)
	}

	// Files should land in ~/Downloads.
	downloadsDir := filepath.Join(home, "Downloads")
	entries, err := os.ReadDir(downloadsDir)
	if err != nil {
		t.Fatalf("reading Downloads dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no skill files exported to ~/Downloads")
	}
}

func TestSetupClaudeOverwrites(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(cwd)
	defer os.Chdir(origDir)

	// First install
	cmd1 := newRootCmd()
	cmd1.SetArgs([]string{"setup", "claude"})
	cmd1.Execute()

	// Second install — should not error
	cmd2 := newRootCmd()
	buf := &bytes.Buffer{}
	cmd2.SetOut(buf)
	cmd2.SetErr(buf)
	cmd2.SetArgs([]string{"setup", "claude"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("second setup claude: %v", err)
	}
}
