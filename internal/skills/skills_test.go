package skills_test

import (
	"strings"
	"testing"

	"github.com/mytours/stqry-cli/internal/skills"
)

func TestHashContent(t *testing.T) {
	h1 := skills.HashContent([]byte("hello"))
	h2 := skills.HashContent([]byte("hello"))
	h3 := skills.HashContent([]byte("world"))

	if h1 != h2 {
		t.Error("same input should produce same hash")
	}
	if h1 == h3 {
		t.Error("different input should produce different hash")
	}
	if len(h1) != 12 {
		t.Errorf("expected 12 hex chars, got %d: %s", len(h1), h1)
	}
}

func TestBuildFrontmatter(t *testing.T) {
	content := []byte("# Hello\nworld\n")
	result := skills.BuildFrontmatter("v1.2.3", content)

	if !strings.HasPrefix(string(result), "---\n") {
		t.Error("expected YAML frontmatter opening ---")
	}
	if !strings.Contains(string(result), "skill_version: v1.2.3") {
		t.Error("expected skill_version in frontmatter")
	}
	if !strings.Contains(string(result), "skill_hash:") {
		t.Error("expected skill_hash in frontmatter")
	}
	if !strings.Contains(string(result), "generated_by: stqry-cli") {
		t.Error("expected generated_by in frontmatter")
	}
	if !strings.Contains(string(result), "# Hello\nworld\n") {
		t.Error("expected original content after frontmatter")
	}
	expectedHash := skills.HashContent(content)
	if !strings.Contains(string(result), "skill_hash: "+expectedHash) {
		t.Errorf("expected skill_hash: %s in frontmatter", expectedHash)
	}
}

func TestExtractHashFromFrontmatter(t *testing.T) {
	content := []byte("---\nskill_version: v1.2.3\nskill_hash: abc123def456\ngenerated_by: stqry-cli\n---\n# Hello\n")
	hash, ok := skills.ExtractHashFromFrontmatter(content)
	if !ok {
		t.Fatal("expected to find hash in frontmatter")
	}
	if hash != "abc123def456" {
		t.Errorf("expected abc123def456, got %s", hash)
	}
}

func TestExtractHashFromFrontmatter_NoFrontmatter(t *testing.T) {
	content := []byte("# Hello\nno frontmatter\n")
	_, ok := skills.ExtractHashFromFrontmatter(content)
	if ok {
		t.Error("expected no hash found for file without frontmatter")
	}
}

func TestEmbeddedSkillNames(t *testing.T) {
	names, err := skills.EmbeddedSkillNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) == 0 {
		t.Error("expected at least one embedded skill")
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".md") {
			t.Errorf("expected .md file, got %s", name)
		}
	}
}
