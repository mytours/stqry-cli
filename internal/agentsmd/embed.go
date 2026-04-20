package agentsmd

import _ "embed"

//go:embed AGENTS.md
var AgentsContent []byte

//go:embed CLAUDE.md
var ClaudeContent []byte
