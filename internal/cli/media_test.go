package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestMediaCreateInvalidType asserts client-side validation of --type on
// `stqry media create`.
func TestMediaCreateInvalidType(t *testing.T) {
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "create", "--type=bogus"})
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid media type, got nil")
	}
	if !contains(err.Error(), "invalid media type") {
		t.Errorf("expected error to mention \"invalid media type\", got %q", err.Error())
	}
	if !contains(err.Error(), "image") {
		t.Errorf("expected error to list valid types (image), got %q", err.Error())
	}
}

// TestMediaListCmd asserts that `stqry media list` prints the media item name
// returned by the API.
func TestMediaListCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/media_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_items": []interface{}{
				map[string]interface{}{"id": "1", "name": "banner", "type": "image"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 1,
			},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "list"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "banner") {
		t.Errorf("expected output to contain %q, got:\n%s", "banner", out)
	}
}

// TestMediaGetCmd asserts that `stqry media get 55` prints the media item name
// returned by the API.
func TestMediaGetCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/media_items/55" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "55", "name": "banner", "type": "image"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "get", "55"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "banner") {
		t.Errorf("expected output to contain %q, got:\n%s", "banner", out)
	}
}

// TestMediaCreateCmd asserts that `stqry media create --type=audio --name="City Tour"`
// sends the correct fields in the POST request body.
func TestMediaCreateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/media_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "99", "name": "City Tour", "type": "audio"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "create", "--type=audio", "--name=City Tour"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["type"] != "audio" {
		t.Errorf("expected type=%q in request body, got %v", "audio", captured["type"])
	}
	if captured["name"] != "City Tour" {
		t.Errorf("expected name=%q in request body, got %v", "City Tour", captured["name"])
	}
}

// TestMediaUpdateCmd asserts that `stqry media update 55 --name="New Banner"`
// sends the correct field in the PATCH request body.
func TestMediaUpdateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/public/media_items/55" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "55", "name": "New Banner", "type": "image"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "update", "55", "--name=New Banner"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["name"] != "New Banner" {
		t.Errorf("expected name=%q in request body, got %v", "New Banner", captured["name"])
	}
}

// TestMediaDeleteCmd asserts that `stqry media delete 55` sends a DELETE
// request to the correct endpoint and prints confirmation.
func TestMediaDeleteCmd(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/public/media_items/55" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		} else {
			deleted = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	os.Stdout = w

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "delete", "55"})
	cmd.SetErr(os.Stderr)
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !deleted {
		t.Error("expected DELETE request to have been made, but it was not")
	}
	if !contains(out, "deleted") {
		t.Errorf("expected output to contain %q, got:\n%s", "deleted", out)
	}
}

// TestMediaCreateCmdImageMetadata asserts that image-specific metadata flags
// (--caption, --attribution, --description) are sent as TranslatedString maps
// keyed by the effective language, so credits can live on the MediaItem at
// create time (instead of being shoved into the section title).
func TestMediaCreateCmdImageMetadata(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/media_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": 42, "type": "image"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--site=testsite", "media", "create",
		"--type=image",
		"--caption=Court Street entrance, c. 1928",
		"--attribution=Rochester Subway Archive · public domain",
		"--description=Vintage postcard of the station entrance",
	})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	for _, field := range []string{"caption", "attribution", "description"} {
		got, ok := captured[field].(map[string]interface{})
		if !ok {
			t.Errorf("expected %s to be a translated-string map, got %v (type %T)", field, captured[field], captured[field])
			continue
		}
		if _, present := got["en"]; !present {
			t.Errorf("expected %s.en to be set, got %v", field, got)
		}
	}
}

// TestMediaCreateCmdOmitsMetadataWhenNotPassed asserts that the new metadata
// fields are NOT sent when their flags are omitted, so existing behaviour is
// preserved for callers that don't opt in.
func TestMediaCreateCmdOmitsMetadataWhenNotPassed(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": 42, "type": "image"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "create", "--type=image"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, field := range []string{"caption", "attribution", "description", "title", "transcription"} {
		if _, present := captured[field]; present {
			t.Errorf("expected no %s field when flag omitted, got %v", field, captured[field])
		}
	}
}

// TestMediaUpdateCmdAudioMetadata asserts that audio-specific flags (--title,
// --transcription) reach the PATCH body as TranslatedString maps, and that
// update uses the Visit() pattern so passing an empty string clears a field.
func TestMediaUpdateCmdAudioMetadata(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/public/media_items/72" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": 72, "type": "audio"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--site=testsite", "media", "update", "72",
		"--title=Stop 1 narration",
		"--transcription=Welcome to the tour...",
	})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	title, ok := captured["title"].(map[string]interface{})
	if !ok || title["en"] != "Stop 1 narration" {
		t.Errorf("expected title.en=\"Stop 1 narration\", got %v", captured["title"])
	}
	trans, ok := captured["transcription"].(map[string]interface{})
	if !ok || trans["en"] != "Welcome to the tour..." {
		t.Errorf("expected transcription.en, got %v", captured["transcription"])
	}
}

// TestMediaUpdateCmdClearFieldWithEmptyString asserts that `--attribution ""`
// sends attribution: {en: ""} (the Visit() pattern), so callers can clear a
// field rather than being stuck with stale credits.
func TestMediaUpdateCmdClearFieldWithEmptyString(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": 55},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "media", "update", "55", "--attribution="})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	attr, ok := captured["attribution"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected attribution map in body, got %v", captured["attribution"])
	}
	if attr["en"] != "" {
		t.Errorf("expected attribution.en=\"\" (cleared), got %v", attr["en"])
	}
}
