package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestScreensCreateInvalidType asserts that `stqry screens create` rejects an
// unknown --type value client-side with a helpful message, before ever hitting
// the API.
func TestScreensCreateInvalidType(t *testing.T) {
	// No server needed: validation should short-circuit.
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "create", "--name=Welcome", "--type=standard"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid screen type, got nil")
	}
	if !contains(err.Error(), "invalid screen type") {
		t.Errorf("expected error to mention \"invalid screen type\", got %q", err.Error())
	}
	if !contains(err.Error(), "story") {
		t.Errorf("expected error to list valid types (story), got %q", err.Error())
	}
}

// TestScreensCreateFlatBody asserts that a valid create actually sends fields
// at the top level (flat), with `type` (not `screen_type`), matching what the
// Rails public API expects.
func TestScreensCreateFlatBody(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/screens" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "1", "name": "Welcome", "type": "story"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "create", "--name=Welcome", "--type=story"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if _, wrapped := captured["screen"]; wrapped {
		t.Errorf("expected flat body, got wrapper: %v", captured)
	}
	if captured["name"] != "Welcome" {
		t.Errorf("expected name=Welcome, got %v", captured["name"])
	}
	if captured["type"] != "story" {
		t.Errorf("expected type=story, got %v", captured["type"])
	}
	if _, legacy := captured["screen_type"]; legacy {
		t.Error("CLI still sending legacy screen_type field")
	}
}

// TestCollectionsCreateInvalidType asserts client-side validation of
// --type on `stqry collections create`.
func TestCollectionsCreateInvalidType(t *testing.T) {
	setupTestHome(t, "http://localhost:0")

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "collections", "create", "--name=test", "--type=bogus"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid collection type, got nil")
	}
	if !contains(err.Error(), "invalid collection type") {
		t.Errorf("expected error to mention \"invalid collection type\", got %q", err.Error())
	}
	if !contains(err.Error(), "tour") {
		t.Errorf("expected error to list valid types (tour), got %q", err.Error())
	}
}

// TestScreensListCmd asserts that `stqry screens list` prints the screen name
// returned by the API.
func TestScreensListCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/screens" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screens": []interface{}{
				map[string]interface{}{"id": "1", "name": "welcome", "type": "story"},
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
	cmd.SetArgs([]string{"--site=testsite", "screens", "list"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "welcome") {
		t.Errorf("expected output to contain %q, got:\n%s", "welcome", out)
	}
}

// TestScreensGetCmd asserts that `stqry screens get 42` prints the screen name
// returned by the API.
func TestScreensGetCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "42", "name": "welcome", "type": "story"},
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
	cmd.SetArgs([]string{"--site=testsite", "screens", "get", "42"})
	execErr := cmd.Execute()

	w.Close()
	os.Stdout = origStdout

	outBytes, _ := io.ReadAll(r)
	r.Close()
	out := string(outBytes)

	if execErr != nil {
		t.Fatalf("Execute: %v", execErr)
	}
	if !contains(out, "welcome") {
		t.Errorf("expected output to contain %q, got:\n%s", "welcome", out)
	}
}

// TestScreensUpdateCmd asserts that `stqry screens update 42 --name=new-name`
// sends the correct field in the PATCH request body.
func TestScreensUpdateCmd(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "42", "name": "new-name", "type": "story"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "update", "42", "--name=new-name"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["name"] != "new-name" {
		t.Errorf("expected name=%q in request body, got %v", "new-name", captured["name"])
	}
}

// TestScreensUpdateCmdHeaderLayout asserts that --header-layout reaches the
// PATCH body as a flat string. This is the one flag that promotes a screen's
// cover image into the actual header, replacing a redundant single_media
// section at the top of each stop.
func TestScreensUpdateCmdHeaderLayout(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": "42", "type": "story", "header_layout": "image_and_title"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "update", "42", "--header-layout=image_and_title"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured["header_layout"] != "image_and_title" {
		t.Errorf("expected header_layout=image_and_title in body, got %v", captured["header_layout"])
	}
}

// TestScreensUpdateCmdInvalidHeaderLayout asserts that --header-layout is
// validated client-side against the enum. A typo like "banner" used to reach
// the API and return a cryptic validation error.
func TestScreensUpdateCmdInvalidHeaderLayout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s %s; should have errored client-side", r.Method, r.URL.Path)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "update", "42", "--header-layout=banner"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid header layout, got nil")
	}
	if !contains(err.Error(), "invalid header layout") {
		t.Errorf("expected error to mention \"invalid header layout\", got %q", err.Error())
	}
	if !contains(err.Error(), "image_and_title") {
		t.Errorf("expected error to list valid layouts, got %q", err.Error())
	}
}

// TestScreensCreateCmdHeaderLayout asserts that --header-layout is accepted at
// create time and sent in the POST body, saving a follow-up update call.
func TestScreensCreateCmdHeaderLayout(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/screens" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"screen": map[string]interface{}{"id": 42, "type": "story"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "create", "--name=Stop 1", "--type=story", "--header-layout=tall"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured["header_layout"] != "tall" {
		t.Errorf("expected header_layout=tall in body, got %v", captured["header_layout"])
	}
}

// TestScreensDeleteCmd asserts that `stqry screens delete 42` sends a DELETE
// request to the correct endpoint.
func TestScreensDeleteCmd(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/api/public/screens/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		} else {
			deleted = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "delete", "42"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !deleted {
		t.Error("expected DELETE request to have been made, but it was not")
	}
}

// TestSectionsUpdateCmdLayout asserts that --layout reaches the PATCH body as
// a flat string. This is Lydia's "change all links to buttons" ask: a
// link_group's layout drives whether links render as list rows or as
// buttons.
func TestSectionsUpdateCmdLayout(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"story_section": map[string]interface{}{"id": 99, "type": "link_group", "layout": "button_with_icon"},
		})
	}))
	defer server.Close()

	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "sections", "update", "99", "--screen-id=42", "--layout=button_with_icon"})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured["layout"] != "button_with_icon" {
		t.Errorf("expected layout=button_with_icon in body, got %v", captured["layout"])
	}
}

// TestSectionsUpdateCmdInvalidLayout asserts that --layout is validated
// client-side against the union of link_group + social_group layouts.
func TestSectionsUpdateCmdInvalidLayout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s %s; should have errored client-side", r.Method, r.URL.Path)
	}))
	defer server.Close()
	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--site=testsite", "screens", "sections", "update", "99", "--screen-id=42", "--layout=pretty"})
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid section layout, got nil")
	}
	if !contains(err.Error(), "invalid section layout") {
		t.Errorf("expected error to mention \"invalid section layout\", got %q", err.Error())
	}
	if !contains(err.Error(), "button_with_icon") {
		t.Errorf("expected error to list valid layouts, got %q", err.Error())
	}
}

// TestLinksAddCmdFieldMapping asserts that `links add` sends the correct
// snake_case API field names (link_type, link_text, stock_icon, icon_type,
// item_type, item_id) and wraps the TranslatedString fields (url, link_text,
// username) in a {lang: value} map. Pre-refactor the command was sending the
// raw flag names ("link-type" etc.) with flat string values for TranslatedString
// fields, so the server was either rejecting or silently ignoring them.
func TestLinksAddCmdFieldMapping(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/screens/42/story_sections/99/link_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{"id": 7, "link_type": "url"},
		})
	}))
	defer server.Close()
	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--site=testsite", "screens", "sections", "links", "add",
		"--screen-id=42", "--section-id=99",
		"--link-type=url",
		"--url=https://example.com",
		"--link-text=Visit site",
		"--icon-type=clear",
	})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["link_type"] != "url" {
		t.Errorf("expected link_type=url (snake_case) in body, got %v; check flag → API field mapping", captured["link_type"])
	}
	if captured["icon_type"] != "clear" {
		t.Errorf("expected icon_type=clear in body, got %v", captured["icon_type"])
	}
	// stock_icon is never configurable from the CLI; it must not appear even
	// when icon_type is explicitly set.
	if _, present := captured["stock_icon"]; present {
		t.Errorf("body should not contain stock_icon (icon is auto-chosen from link_type), got %v", captured["stock_icon"])
	}
	url, ok := captured["url"].(map[string]interface{})
	if !ok || url["en"] != "https://example.com" {
		t.Errorf("expected url={en: https://example.com} (TranslatedString), got %v", captured["url"])
	}
	linkText, ok := captured["link_text"].(map[string]interface{})
	if !ok || linkText["en"] != "Visit site" {
		t.Errorf("expected link_text={en: Visit site} (TranslatedString), got %v", captured["link_text"])
	}
	// Old flag name must not leak through.
	if _, bad := captured["link-type"]; bad {
		t.Errorf("body should not contain kebab-case key \"link-type\"; got %v", captured["link-type"])
	}
}

// TestLinksAddCmdInternalTarget asserts that --link-type=internal with
// --item-type=Screen --item-id=100 sends item_type + item_id as snake_case
// fields with item_id as an integer (not a string).
func TestLinksAddCmdInternalTarget(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{"id": 7},
		})
	}))
	defer server.Close()
	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--site=testsite", "screens", "sections", "links", "add",
		"--screen-id=42", "--section-id=99",
		"--link-type=internal",
		"--item-type=Screen",
		"--item-id=100",
		"--link-text=Related stop",
	})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["item_type"] != "Screen" {
		t.Errorf("expected item_type=Screen, got %v", captured["item_type"])
	}
	id, ok := captured["item_id"].(float64)
	if !ok || int(id) != 100 {
		t.Errorf("expected item_id=100 as integer, got %v (type %T)", captured["item_id"], captured["item_id"])
	}
}

// TestLinksAddCmdSocialUsername asserts that --link-type=instagram with
// --username sends username as a TranslatedString. Covers the "social links
// capability" part of Lydia's request (social link types live on link_items,
// keyed by username rather than a full URL).
func TestLinksAddCmdSocialUsername(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{"id": 7},
		})
	}))
	defer server.Close()
	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--site=testsite", "screens", "sections", "links", "add",
		"--screen-id=42", "--section-id=99",
		"--link-type=instagram",
		"--username=@example",
		"--link-text=Follow us",
	})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["link_type"] != "instagram" {
		t.Errorf("expected link_type=instagram, got %v", captured["link_type"])
	}
	username, ok := captured["username"].(map[string]interface{})
	if !ok || username["en"] != "@example" {
		t.Errorf("expected username={en: @example} (TranslatedString), got %v", captured["username"])
	}
}

// TestLinksAddCmdInvalidEnums asserts that --link-type, --icon-type, and
// --item-type are each validated client-side against their enum.
func TestLinksAddCmdInvalidEnums(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s %s; should have errored client-side", r.Method, r.URL.Path)
	}))
	defer server.Close()
	setupTestHome(t, server.URL)

	cases := []struct {
		name    string
		flags   []string
		wantErr string
	}{
		{"link-type", []string{"--link-type=bogus"}, "invalid link type"},
		{"icon-type", []string{"--link-type=url", "--icon-type=fancy"}, "invalid icon type"},
		{"item-type", []string{"--link-type=internal", "--item-type=Widget", "--item-id=1"}, "invalid item type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{
				"--site=testsite", "screens", "sections", "links", "add",
				"--screen-id=42", "--section-id=99",
			}, tc.flags...)
			cmd := newRootCmd()
			cmd.SetArgs(args)
			cmd.SetOut(os.Stderr)
			cmd.SetErr(os.Stderr)
			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !contains(err.Error(), tc.wantErr) {
				t.Errorf("expected error to mention %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

// TestSocialAddCmdFieldMapping asserts that the social sub-item add sends
// social_network (snake_case), wraps username and link_text as
// TranslatedString, and does NOT carry the stale --url flag the command used
// to advertise (social items live on a separate resource keyed by username,
// not a URL).
func TestSocialAddCmdFieldMapping(t *testing.T) {
	var captured map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/public/screens/42/story_sections/99/social_items" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"item": map[string]interface{}{"id": 7, "social_network": "instagram"},
		})
	}))
	defer server.Close()
	setupTestHome(t, server.URL)

	cmd := newRootCmd()
	cmd.SetArgs([]string{
		"--site=testsite", "screens", "sections", "social", "add",
		"--screen-id=42", "--section-id=99",
		"--social-network=instagram",
		"--username=@example",
		"--link-text=Follow us",
	})
	cmd.SetErr(os.Stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if captured["social_network"] != "instagram" {
		t.Errorf("expected social_network=instagram (snake_case), got %v", captured["social_network"])
	}
	username, ok := captured["username"].(map[string]interface{})
	if !ok || username["en"] != "@example" {
		t.Errorf("expected username={en: @example}, got %v", captured["username"])
	}
	linkText, ok := captured["link_text"].(map[string]interface{})
	if !ok || linkText["en"] != "Follow us" {
		t.Errorf("expected link_text={en: Follow us}, got %v", captured["link_text"])
	}
}
