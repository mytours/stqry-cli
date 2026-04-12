package mcp_test

import (
	"sync"
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

func TestSessionClear(t *testing.T) {
	sess := stqrymcp.NewSession()
	sess.Set(&config.Site{Token: "tok", APIURL: "https://api.example.com"})
	if sess.Get() == nil {
		t.Fatal("expected site after Set")
	}
	sess.Clear()
	if sess.Get() != nil {
		t.Error("expected nil site after Clear")
	}
}

func TestSessionConcurrent(t *testing.T) {
	sess := stqrymcp.NewSession()
	site := &config.Site{Token: "tok", APIURL: "https://api.example.com"}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); sess.Set(site) }()
		go func() { defer wg.Done(); _ = sess.Get() }()
	}
	wg.Wait()
}
