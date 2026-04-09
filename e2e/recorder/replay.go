package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type replayProxy struct {
	interactions []Interaction
	mu           sync.Mutex
}

func newReplayProxy(cassettesDir string) (*replayProxy, error) {
	entries, err := os.ReadDir(cassettesDir)
	if err != nil {
		return nil, fmt.Errorf("reading cassettes dir %q: %w", cassettesDir, err)
	}

	var interactions []Interaction
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(cassettesDir, entry.Name())
		c, err := loadCassette(path)
		if err != nil {
			return nil, fmt.Errorf("loading cassette %q: %w", path, err)
		}
		interactions = append(interactions, c.Interactions...)
	}

	fmt.Printf("Loaded %d interactions from %q\n", len(interactions), cassettesDir)
	return &replayProxy{interactions: interactions}, nil
}

func (p *replayProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Build canonical key using path+query (not full URL) so host-independent matching works.
	reqURL := r.URL.RequestURI()
	key := canonicalKey(r.Method, reqURL, string(bodyBytes))

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, interaction := range p.interactions {
		if interaction.Request.CanonicalKey == key {
			// Write recorded headers.
			for k, v := range interaction.Response.Headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(interaction.Response.Status)
			_, _ = w.Write([]byte(interaction.Response.Body))
			return
		}
	}

	// No match found — return a descriptive 502.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	resp := map[string]string{
		"error": fmt.Sprintf("no cassette interaction matched: %s", key),
	}
	_ = json.NewEncoder(w).Encode(resp)
}
