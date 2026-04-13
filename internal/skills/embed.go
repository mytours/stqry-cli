package skills

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"strings"
)

//go:embed *.md
var SkillFiles embed.FS

// HashContent returns the first 12 hex chars of the SHA256 of data.
func HashContent(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:6])
}

// BuildFrontmatter prepends YAML frontmatter to content and returns the combined bytes.
// The skill_hash is computed from content (without frontmatter).
// If content already begins with a YAML frontmatter block (e.g. name/description fields),
// the two blocks are merged into one so the output has a single --- … --- header.
func BuildFrontmatter(version string, content []byte) []byte {
	hash := HashContent(content)
	newFields := fmt.Sprintf("skill_version: %s\nskill_hash: %s\ngenerated_by: stqry-cli\n", version, hash)

	s := string(content)
	if strings.HasPrefix(s, "---\n") {
		rest := s[4:]
		if end := strings.Index(rest, "\n---\n"); end >= 0 {
			existingFields := rest[:end+1] // existing fields with trailing newline
			body := rest[end+5:]           // content after closing ---
			return []byte("---\n" + newFields + existingFields + "---\n" + body)
		}
	}

	fm := fmt.Sprintf("---\n%s---\n", newFields)
	return append([]byte(fm), content...)
}

// ExtractHashFromFrontmatter reads the skill_hash field from YAML frontmatter.
// Returns ("", false) if the file has no frontmatter or no skill_hash field.
func ExtractHashFromFrontmatter(data []byte) (string, bool) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return "", false
	}
	rest := s[4:]
	// Normalise: ensure content ends with newline so "\n---\n" always matches.
	if !strings.HasSuffix(rest, "\n") {
		rest += "\n"
	}
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return "", false
	}
	block := rest[:end]
	for _, line := range strings.Split(block, "\n") {
		if strings.HasPrefix(line, "skill_hash: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "skill_hash: ")), true
		}
	}
	return "", false
}

// EmbeddedSkillNames returns the filenames of all embedded skill files.
func EmbeddedSkillNames() ([]string, error) {
	entries, err := SkillFiles.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}
