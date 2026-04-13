package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestSkillDumpListsSkills(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"skill", "dump"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("skill dump: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "stqry-reference") {
		t.Errorf("expected stqry-reference in listing, got: %s", out)
	}
}

func TestSkillDumpNamedSkill(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"skill", "dump", "stqry-reference"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("skill dump stqry-reference: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "---\n") {
		t.Errorf("expected frontmatter at start, got: %s", out[:min(100, len(out))])
	}
	if !strings.Contains(out, "skill_hash:") {
		t.Error("expected skill_hash in output")
	}
}

func TestSkillDumpUnknownSkill(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"skill", "dump", "no-such-skill"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for unknown skill name")
	}
}
