package mcp

import (
	"sync"

	"github.com/mytours/stqry-cli/internal/config"
)

// Session holds in-memory site credentials for the duration of the MCP server process.
type Session struct {
	mu   sync.Mutex
	site *config.Site
}

// NewSession returns an empty session.
func NewSession() *Session {
	return &Session{}
}

// Set stores the active site. Safe for concurrent use.
func (s *Session) Set(site *config.Site) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.site = site
}

// Get returns the active site, or nil if none is set.
func (s *Session) Get() *config.Site {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.site
}
