package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(got, "stqry ") {
		t.Errorf("expected version output to start with \"stqry \", got %q", got)
	}
}

func TestVersionShortFlag(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"-v"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := strings.TrimSpace(buf.String())
	if !strings.HasPrefix(got, "stqry ") {
		t.Errorf("expected version output to start with \"stqry \", got %q", got)
	}
}

func TestCompletionDoesNotRequireSite(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"completion", "zsh"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("completion zsh should not require a site, got: %v", err)
	}
}
