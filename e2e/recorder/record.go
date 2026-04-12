package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type recordProxy struct {
	target      *url.URL
	cassette    *Cassette
	mu          sync.Mutex
	cassettePath string
}

func newRecordProxy(target string, cassettesDir string) (*recordProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL %q: %w", target, err)
	}

	filename := fmt.Sprintf("recording-%s.json", time.Now().Format("20060102-150405"))
	cassettePath := filepath.Join(cassettesDir, filename)

	return &recordProxy{
		target:       u,
		cassette:     &Cassette{},
		cassettePath: cassettePath,
	}, nil
}

func (p *recordProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read request body.
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Build the upstream URL.
	upstream := *p.target
	upstream.Path = strings.TrimRight(upstream.Path, "/") + r.URL.Path
	upstream.RawQuery = r.URL.RawQuery

	// Create the proxied request.
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, upstream.String(), strings.NewReader(string(reqBody)))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create upstream request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy headers (skip hop-by-hop headers).
	for key, values := range r.Header {
		for _, v := range values {
			proxyReq.Header.Add(key, v)
		}
	}
	proxyReq.Header.Del("Host")

	// Execute the upstream request. TLS verification is skipped so that local
	// dev servers with self-signed or mkcert certificates work without extra setup.
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
	resp, err := client.Do(proxyReq)
	if err != nil {
		fmt.Printf("Upstream error: %s %s → %v\n", r.Method, upstream.String(), err)
		http.Error(w, fmt.Sprintf("upstream request failed: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read upstream response body", http.StatusInternalServerError)
		return
	}

	// Flatten response headers (first value only, sufficient for cassettes).
	respHeaders := make(map[string]string, len(resp.Header))
	for k, vs := range resp.Header {
		if len(vs) > 0 {
			respHeaders[k] = vs[0]
		}
	}

	// Build canonical key using path+query so replays are host-independent.
	reqURL := r.URL.RequestURI()
	key := canonicalKey(r.Method, reqURL, string(reqBody))

	interaction := Interaction{
		Request: CassetteRequest{
			Method:       r.Method,
			URL:          upstream.String(),
			Body:         string(reqBody),
			CanonicalKey: key,
		},
		Response: CassetteResponse{
			Status:  resp.StatusCode,
			Headers: respHeaders,
			Body:    string(respBody),
		},
	}

	p.mu.Lock()
	p.cassette.Interactions = append(p.cassette.Interactions, interaction)
	p.mu.Unlock()

	// Forward the response to the caller.
	for k, v := range respHeaders {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)

	fmt.Printf("Recorded: %s %s → %d\n", r.Method, reqURL, resp.StatusCode)
}

// save writes the cassette to disk.
func (p *recordProxy) save() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.cassette.Interactions) == 0 {
		fmt.Println("No interactions recorded, skipping cassette save.")
		return nil
	}

	if err := saveCassette(p.cassettePath, p.cassette); err != nil {
		return fmt.Errorf("saving cassette to %q: %w", p.cassettePath, err)
	}
	fmt.Printf("Saved %d interactions to %q\n", len(p.cassette.Interactions), p.cassettePath)
	return nil
}
