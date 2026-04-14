package agentsmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/mytours/stqry-cli/internal/agentsmd"
)

func TestAgentsMDInSync(t *testing.T) {
	root, err := os.ReadFile("../../AGENTS.md")
	if err != nil {
		t.Fatalf("reading root AGENTS.md: %v", err)
	}
	if !bytes.Equal(root, agentsmd.Content) {
		t.Fatal("root AGENTS.md and internal/agentsmd/AGENTS.md are out of sync — update both together")
	}
}
