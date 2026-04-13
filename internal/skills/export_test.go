package skills_test

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/mytours/stqry-cli/internal/skills"
)

func TestBuildZip_ContainsRequiredFiles(t *testing.T) {
	data, err := skills.BuildZip("v1.2.3")
	if err != nil {
		t.Fatalf("BuildZip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opening zip: %v", err)
	}

	want := []string{
		"stqry-skill/SKILL.md",
		"stqry-skill/SETUP.md",
		"stqry-skill/REFERENCE.md",
		"stqry-skill/WORKFLOWS.md",
	}
	found := map[string]bool{}
	for _, f := range r.File {
		found[f.Name] = true
	}
	for _, name := range want {
		if !found[name] {
			t.Errorf("zip missing expected file: %s", name)
		}
	}
	if len(r.File) != len(want) {
		t.Errorf("expected %d files in zip, got %d", len(want), len(r.File))
	}
}

func TestBuildZip_SkillMDHasFrontmatterAndVersion(t *testing.T) {
	data, err := skills.BuildZip("v1.2.3")
	if err != nil {
		t.Fatalf("BuildZip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opening zip: %v", err)
	}

	for _, f := range r.File {
		if f.Name != "stqry-skill/SKILL.md" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening SKILL.md: %v", err)
		}
		var buf bytes.Buffer
		buf.ReadFrom(rc)
		rc.Close()

		content := buf.String()
		if !strings.HasPrefix(content, "---\n") {
			t.Error("SKILL.md: missing frontmatter")
		}
		if !strings.Contains(content, "skill_version: v1.2.3") {
			t.Error("SKILL.md: missing skill_version")
		}
		if !strings.Contains(content, "name: stqry") {
			t.Error("SKILL.md: missing name field")
		}
		return
	}
	t.Error("SKILL.md not found in zip")
}

func TestBuildZip_ReferenceMDHasFrontmatter(t *testing.T) {
	data, err := skills.BuildZip("v1.2.3")
	if err != nil {
		t.Fatalf("BuildZip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opening zip: %v", err)
	}

	for _, f := range r.File {
		if f.Name != "stqry-skill/REFERENCE.md" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening REFERENCE.md: %v", err)
		}
		var buf bytes.Buffer
		buf.ReadFrom(rc)
		rc.Close()

		content := buf.String()
		if !strings.HasPrefix(content, "---\n") {
			t.Error("REFERENCE.md: missing frontmatter")
		}
		if !strings.Contains(content, "skill_version: v1.2.3") {
			t.Error("REFERENCE.md: missing skill_version")
		}
		return
	}
	t.Error("REFERENCE.md not found in zip")
}

func TestBuildZip_SetupMDContainsInstallGuidance(t *testing.T) {
	data, err := skills.BuildZip("v1.2.3")
	if err != nil {
		t.Fatalf("BuildZip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opening zip: %v", err)
	}

	for _, f := range r.File {
		if f.Name != "stqry-skill/SETUP.md" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening SETUP.md: %v", err)
		}
		var buf bytes.Buffer
		buf.ReadFrom(rc)
		rc.Close()

		content := buf.String()
		if !strings.Contains(content, "pip install stqry") {
			t.Error("SETUP.md: missing pip install stqry")
		}
		if !strings.Contains(content, "stqry config add-site") {
			t.Error("SETUP.md: missing config add-site")
		}
		if !strings.Contains(content, "stqry.yaml") {
			t.Error("SETUP.md: missing stqry.yaml persistence guidance")
		}
		return
	}
	t.Error("SETUP.md not found in zip")
}

func TestBuildZip_DifferentVersionsProduceDifferentContent(t *testing.T) {
	d1, err := skills.BuildZip("v1.0.0")
	if err != nil {
		t.Fatalf("BuildZip v1.0.0: %v", err)
	}
	d2, err := skills.BuildZip("v2.0.0")
	if err != nil {
		t.Fatalf("BuildZip v2.0.0: %v", err)
	}

	r1, err := zip.NewReader(bytes.NewReader(d1), int64(len(d1)))
	if err != nil {
		t.Fatalf("opening zip v1.0.0: %v", err)
	}
	r2, err := zip.NewReader(bytes.NewReader(d2), int64(len(d2)))
	if err != nil {
		t.Fatalf("opening zip v2.0.0: %v", err)
	}

	readFile := func(r *zip.Reader, name string) string {
		for _, f := range r.File {
			if f.Name == name {
				rc, _ := f.Open()
				var buf bytes.Buffer
				buf.ReadFrom(rc)
				rc.Close()
				return buf.String()
			}
		}
		return ""
	}

	s1 := readFile(r1, "stqry-skill/SKILL.md")
	s2 := readFile(r2, "stqry-skill/SKILL.md")
	if s1 == s2 {
		t.Error("expected different SKILL.md content for different versions")
	}
	if !strings.Contains(s1, "v1.0.0") {
		t.Error("v1.0.0 SKILL.md should contain v1.0.0")
	}
	if !strings.Contains(s2, "v2.0.0") {
		t.Error("v2.0.0 SKILL.md should contain v2.0.0")
	}
}

func TestBuildZip_WorkflowsMDHasFrontmatter(t *testing.T) {
	data, err := skills.BuildZip("v1.2.3")
	if err != nil {
		t.Fatalf("BuildZip: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("opening zip: %v", err)
	}

	for _, f := range r.File {
		if f.Name != "stqry-skill/WORKFLOWS.md" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening WORKFLOWS.md: %v", err)
		}
		var buf bytes.Buffer
		buf.ReadFrom(rc)
		rc.Close()

		content := buf.String()
		if !strings.HasPrefix(content, "---\n") {
			t.Error("WORKFLOWS.md: missing frontmatter")
		}
		if !strings.Contains(content, "skill_version: v1.2.3") {
			t.Error("WORKFLOWS.md: missing skill_version")
		}
		return
	}
	t.Error("WORKFLOWS.md not found in zip")
}
