package cli

import (
	"bytes"
	"testing"
)

func TestCompletionBashCmd(t *testing.T) {
	setupTestHome(t, "http://unused")

	buf := &bytes.Buffer{}
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "bash"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("bash")) {
		t.Error("expected bash completion script in output")
	}
}

func TestCompletionZshCmd(t *testing.T) {
	setupTestHome(t, "http://unused")

	buf := &bytes.Buffer{}
	cmd := newRootCmd()
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "zsh"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("zsh")) {
		t.Error("expected zsh completion script in output")
	}
}
