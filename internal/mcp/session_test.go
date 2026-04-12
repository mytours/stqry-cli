package mcp_test

import (
	"testing"

	"github.com/mytours/stqry-cli/internal/config"
	stqrymcp "github.com/mytours/stqry-cli/internal/mcp"
)

func TestSessionGetSetNil(t *testing.T) {
	sess := stqrymcp.NewSession()
	if sess.Get() != nil {
		t.Error("expected nil site on new session")
	}
}

func TestSessionGetSetSite(t *testing.T) {
	sess := stqrymcp.NewSession()
	site := &config.Site{Token: "tok", APIURL: "https://api.example.com"}
	sess.Set(site)
	got := sess.Get()
	if got == nil {
		t.Fatal("expected non-nil site after Set")
	}
	if got.Token != "tok" {
		t.Errorf("expected token tok, got %s", got.Token)
	}
}
