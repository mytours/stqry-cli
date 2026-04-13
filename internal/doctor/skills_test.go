package doctor_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mytours/stqry-cli/internal/doctor"
	"github.com/mytours/stqry-cli/internal/skills"
)

func TestCheckInstalledSkills_Pass(t *testing.T) {
	dir := t.TempDir()
	if err := skills.InstallAll(dir, skills.LayoutCode, "v1.0.0"); err != nil {
		t.Fatalf("InstallAll: %v", err)
	}

	loc := doctor.SkillLocation{Dir: dir, Layout: doctor.SkillLayoutCode, Label: "test (local)"}
	results := doctor.CheckInstalledSkills([]doctor.SkillLocation{loc})

	for _, r := range results {
		if r.Status != doctor.StatusPass {
			t.Errorf("expected pass for %s, got %s: %s", r.Name, r.Status, r.Message)
		}
	}
}

func TestCheckInstalledSkills_Stale(t *testing.T) {
	dir := t.TempDir()

	// Write skill files with a deliberately wrong hash.
	staleContent := []byte("---\nskill_version: v0.1.0\nskill_hash: 000000000000\ngenerated_by: stqry-cli\n---\n# stale\n")
	skillNames, err := skills.EmbeddedSkillNames()
	if err != nil {
		t.Fatalf("getting skill names: %v", err)
	}
	for _, name := range skillNames {
		if err := os.WriteFile(filepath.Join(dir, name), staleContent, 0644); err != nil {
			t.Fatalf("writing stale skill %s: %v", name, err)
		}
	}

	loc := doctor.SkillLocation{Dir: dir, Layout: doctor.SkillLayoutCode, Label: "test (local)"}
	results := doctor.CheckInstalledSkills([]doctor.SkillLocation{loc})

	warned := false
	for _, r := range results {
		if r.Status == doctor.StatusWarn {
			warned = true
			if !strings.Contains(r.Message, "outdated") {
				t.Errorf("expected 'outdated' in message, got: %s", r.Message)
			}
			if !strings.Contains(r.Detail, "stqry setup claude") {
				t.Errorf("expected remediation hint in detail, got: %s", r.Detail)
			}
		}
	}
	if !warned {
		t.Error("expected at least one warn result for stale skill")
	}
}

func TestCheckInstalledSkills_NotInstalled(t *testing.T) {
	dir := t.TempDir() // empty — no skill files

	loc := doctor.SkillLocation{Dir: dir, Layout: doctor.SkillLayoutCode, Label: "test (local)"}
	results := doctor.CheckInstalledSkills([]doctor.SkillLocation{loc})

	for _, r := range results {
		if r.Status != doctor.StatusSkip {
			t.Errorf("expected skip for uninstalled skill %s, got %s", r.Name, r.Status)
		}
	}
}

func TestCheckInstalledSkills_DesktopLayout(t *testing.T) {
	dir := t.TempDir()
	if err := skills.InstallAll(dir, skills.LayoutDesktop, "v1.0.0"); err != nil {
		t.Fatalf("InstallAll: %v", err)
	}

	loc := doctor.SkillLocation{Dir: dir, Layout: doctor.SkillLayoutDesktop, Label: "Claude Desktop"}
	results := doctor.CheckInstalledSkills([]doctor.SkillLocation{loc})

	for _, r := range results {
		if r.Status != doctor.StatusPass {
			t.Errorf("expected pass for %s, got %s: %s", r.Name, r.Status, r.Message)
		}
	}
}
