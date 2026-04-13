package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Layout describes where and how skills are written on disk.
type Layout int

const (
	LayoutCode    Layout = iota // Flat .md file: {dir}/{skill-name}.md
	LayoutDesktop               // Folder+file: {dir}/{skill-name}/SKILL.md
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
		if err := writeSkill(targetDir, name, layout, content); err != nil {
			return err
		}
	}
	return nil
}

// writeSkill writes a single skill's content to disk according to layout.
func writeSkill(targetDir, filename string, layout Layout, content []byte) error {
	skillName := strings.TrimSuffix(filename, ".md")
	var destPath string
	switch layout {
	case LayoutDesktop:
		skillDir := filepath.Join(targetDir, skillName)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			return fmt.Errorf("creating skill directory %s: %w", skillDir, err)
		}
		destPath = filepath.Join(skillDir, "SKILL.md")
	default: // LayoutCode
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", targetDir, err)
		}
		destPath = filepath.Join(targetDir, filename)
	}
	if err := os.WriteFile(destPath, content, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", destPath, err)
	}
	return nil
}

// DesktopSkillsDir returns the OS-appropriate Claude Desktop skills directory.
func DesktopSkillsDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "Claude", "skills")
	case "linux":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "Claude", "skills")
	default: // macOS
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "Claude", "skills")
	}
}
