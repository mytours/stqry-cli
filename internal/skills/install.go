package skills

import (
	"fmt"
	"os"
	"path/filepath"
)

// Layout describes where and how skills are written on disk.
type Layout int

const (
	LayoutCode Layout = iota // Flat .md file: {dir}/{skill-name}.md
)

// InstallAll writes all embedded skills to targetDir using the given layout.
// Always overwrites existing files.
func InstallAll(targetDir string, layout Layout, version string) error {
	names, err := EmbeddedSkillNames()
	if err != nil {
		return fmt.Errorf("reading embedded skills: %w", err)
	}
	for _, name := range names {
		data, err := SkillFiles.ReadFile(name)
		if err != nil {
			return fmt.Errorf("reading embedded file %s: %w", name, err)
		}
		content := BuildFrontmatter(version, data)
		if err := writeSkill(targetDir, name, content); err != nil {
			return err
		}
	}
	return nil
}

// writeSkill writes a single skill's content to disk.
func writeSkill(targetDir, filename string, content []byte) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", targetDir, err)
	}
	destPath := filepath.Join(targetDir, filename)
	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", destPath, err)
	}
	return nil
}
