package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mytours/stqry-cli/internal/skills"
)

// SkillLocation describes a directory where skills may be installed.
type SkillLocation struct {
	Dir   string
	Label string // e.g. "Claude Code (global)"
}

// CheckInstalledSkills checks every embedded skill at each location.
// Returns one CheckResult per skill per location.
func CheckInstalledSkills(locations []SkillLocation) []CheckResult {
	skillNames, err := skills.EmbeddedSkillNames()
	if err != nil {
		return []CheckResult{{
			Group:   "Skills",
			Name:    "Embedded skills",
			Status:  StatusFail,
			Message: fmt.Sprintf("Could not read embedded skills: %v", err),
		}}
	}

	var results []CheckResult
	for _, loc := range locations {
		for _, filename := range skillNames {
			results = append(results, checkOneSkill(loc, filename))
		}
	}
	return results
}

func checkOneSkill(loc SkillLocation, filename string) CheckResult {
	start := time.Now()
	skillName := strings.TrimSuffix(filename, ".md")
	r := CheckResult{
		Group: "Skills",
		Name:  fmt.Sprintf("%s — %s", skillName, loc.Label),
	}

	if loc.Dir == "" {
		r.Status = StatusSkip
		r.Message = "location not available on this platform"
		r.Duration = time.Since(start)
		return r
	}

	installedPath := filepath.Join(loc.Dir, filename)

	data, err := os.ReadFile(installedPath)
	if err != nil {
		r.Status = StatusSkip
		r.Message = "not installed"
		r.Duration = time.Since(start)
		return r
	}

	installedHash, ok := skills.ExtractHashFromFrontmatter(data)
	if !ok {
		r.Status = StatusWarn
		r.Message = "outdated (no version metadata)"
		r.Detail = remediation()
		r.Duration = time.Since(start)
		return r
	}

	embeddedData, err := skills.SkillFiles.ReadFile(filename)
	if err != nil {
		r.Status = StatusFail
		r.Message = "embedded skill missing: " + filename
		r.Duration = time.Since(start)
		return r
	}

	currentHash := skills.HashContent(embeddedData)
	if installedHash != currentHash {
		r.Status = StatusWarn
		r.Message = "outdated (skill content has changed)"
		r.Detail = remediation()
		r.Duration = time.Since(start)
		return r
	}

	r.Status = StatusPass
	r.Message = "up to date"
	r.Duration = time.Since(start)
	return r
}

func remediation() string {
	return "Run: stqry setup claude (or --global)"
}
