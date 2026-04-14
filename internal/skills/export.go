package skills

import (
	"archive/zip"
	"bytes"
	"fmt"
)

// skillMDContent is the instruction body for SKILL.md.
// It includes its own YAML frontmatter (name, description) which BuildFrontmatter
// will merge with the injected fields (skill_version, skill_hash, generated_by)
// into a single frontmatter block. The content must begin with "---\n" and contain
// a valid closing "\n---\n" sequence for the merge to work correctly.
const skillMDContent = "---\n" +
	"name: stqry\n" +
	"description: STQRY CLI — manage STQRY content from Claude Desktop and Claude Cowork\n" +
	"---\n\n" +
	"# STQRY Skill\n\n" +
	"This skill gives you full access to the STQRY CLI for managing tours, collections,\n" +
	"screens, media, and codes.\n\n" +
	"## On Every Use\n\n" +
	"Before any content operation:\n\n" +
	"1. Run `stqry --version` to confirm the CLI is installed and capture its version.\n" +
	"   - If not found: follow the installation steps in SETUP.md.\n" +
	"   - If the installed version is **newer** than the `skill_version` in this file's\n" +
	"     frontmatter, warn the user: \"Your stqry CLI is newer than this skill. Download\n" +
	"     the latest skill from https://github.com/mytours/stqry-cli/releases for updated\n" +
	"     commands and workflows.\"\n\n" +
	"2. Run `stqry doctor` to verify site connectivity.\n" +
	"   - If no site is configured: follow the configuration steps in SETUP.md.\n\n" +
	"Then use REFERENCE.md for command reference and WORKFLOWS.md for step-by-step recipes.\n"

// setupMDContent is the full installation and configuration guide for SETUP.md.
// NOTE: The "Setup & Installation" section in stqry-reference.md mirrors this content.
// Keep both in sync when updating installation or configuration instructions.
const setupMDContent = "# STQRY Setup Guide\n\n" +
	"## Installation\n\n" +
	"### Claude Cowork\n\n" +
	"If `stqry` is not on PATH, install it via pip:\n\n" +
	"    pip install stqry\n\n" +
	"Verify with `stqry --version` before proceeding.\n\n" +
	"Note: MCP server setup is not available in Claude Cowork.\n" +
	"Use CLI subprocess for all operations.\n\n" +
	"### Claude Desktop\n\n" +
	"Install via Homebrew:\n\n" +
	"    brew install mytours/tap/stqry-cli\n\n" +
	"Or download from https://github.com/mytours/stqry-cli/releases.\n\n" +
	"## Site Configuration\n\n" +
	"Add a site to global config:\n\n" +
	"    stqry config add-site --name mysite --token <token> --region us\n\n" +
	"Available regions: us, ca, eu, sg, au\n\n" +
	"Write a project config file to the current directory:\n\n" +
	"    stqry config init --name mysite   # writes stqry.yaml\n\n" +
	"Verify connectivity:\n\n" +
	"    stqry doctor\n\n" +
	"## Persisting Configuration in Claude Cowork\n\n" +
	"After running `stqry config init`, a `stqry.yaml` file is written to the current\n" +
	"directory. To avoid reconfiguring each session:\n\n" +
	"1. Ask Claude to display the contents of `stqry.yaml`\n" +
	"2. Save it to your local machine\n" +
	"3. Upload it at the start of future Cowork sessions, or add it to your Cowork\n" +
	"   working folder so it is picked up automatically\n"

// BuildZip returns the bytes of a zip archive containing the STQRY skill package.
// The zip has the following layout:
//
//	stqry-skill/SKILL.md
//	stqry-skill/SETUP.md
//	stqry-skill/REFERENCE.md
//	stqry-skill/WORKFLOWS.md
func BuildZip(version string) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// SKILL.md — generated with skill_version merged into frontmatter
	if err := addZipEntry(w, "stqry-skill/SKILL.md", BuildFrontmatter(version, []byte(skillMDContent))); err != nil {
		return nil, err
	}

	// SETUP.md — static template, no frontmatter needed
	if err := addZipEntry(w, "stqry-skill/SETUP.md", []byte(setupMDContent)); err != nil {
		return nil, err
	}

	// REFERENCE.md — from embedded stqry-reference.md with frontmatter
	refData, err := SkillFiles.ReadFile("stqry-reference.md")
	if err != nil {
		return nil, fmt.Errorf("reading stqry-reference.md: %w", err)
	}
	if err := addZipEntry(w, "stqry-skill/REFERENCE.md", BuildFrontmatter(version, refData)); err != nil {
		return nil, err
	}

	// WORKFLOWS.md — from embedded stqry-workflows.md with frontmatter
	wfData, err := SkillFiles.ReadFile("stqry-workflows.md")
	if err != nil {
		return nil, fmt.Errorf("reading stqry-workflows.md: %w", err)
	}
	if err := addZipEntry(w, "stqry-skill/WORKFLOWS.md", BuildFrontmatter(version, wfData)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("closing zip: %w", err)
	}
	return buf.Bytes(), nil
}

func addZipEntry(w *zip.Writer, name string, content []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return fmt.Errorf("creating zip entry %s: %w", name, err)
	}
	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("writing zip entry %s: %w", name, err)
	}
	return nil
}
