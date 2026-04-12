package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/config"
	stqrymcp "github.com/mytours/stqry-cli/internal/mcp"
)

// callTool sends a tools/call JSON-RPC message to the server and returns the result.
func callTool(s *mcpserver.MCPServer, name string, args string) *mcpgo.CallToolResult {
	msg := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, name, args)
	resp := s.HandleMessage(context.Background(), json.RawMessage(msg))
	jsonResp, ok := resp.(mcpgo.JSONRPCResponse)
	if !ok {
		return nil
	}
	result, ok := jsonResp.Result.(*mcpgo.CallToolResult)
	if !ok {
		return nil
	}
	return result
}

// toolText returns the text content from the first content item of a CallToolResult.
func toolText(r *mcpgo.CallToolResult) string {
	if r == nil || len(r.Content) == 0 {
		return ""
	}
	text, ok := r.Content[0].(mcpgo.TextContent)
	if !ok {
		return ""
	}
	return text.Text
}

// ---- configure_project ----

func TestConfigureProjectTool(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	err := stqrymcp.WriteProjectConfig("https://api.example.com", "tok123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "stqry.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("tok123")) {
		t.Error("expected token in stqry.yaml")
	}
	if !bytes.Contains(data, []byte("api.example.com")) {
		t.Error("expected api_url in stqry.yaml")
	}
}

func TestConfigureProjectViaHandleMessage(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "configure_project", `{"api_url":"https://api.example.com","token":"tok123"}`)
	if result == nil {
		t.Fatal("expected a result")
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", toolText(result))
	}
	data, err := os.ReadFile(filepath.Join(dir, "stqry.yaml"))
	if err != nil {
		t.Fatal("stqry.yaml not written")
	}
	if !bytes.Contains(data, []byte("tok123")) {
		t.Error("expected token in stqry.yaml")
	}
}

func TestConfigureProjectMissingParams(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "configure_project", `{"api_url":"","token":""}`)
	if result == nil {
		t.Fatal("expected a result")
	}
	if !result.IsError {
		t.Fatal("expected an error result for missing params")
	}
}

func TestConfigureProjectInvalidURL(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")

	for _, badURL := range []string{"not-a-url", "ftp://wrong-scheme.com", "http://"} {
		result := callTool(s, "configure_project", fmt.Sprintf(`{"api_url":%q,"token":"tok"}`, badURL))
		if result == nil {
			t.Fatalf("expected a result for URL %q", badURL)
		}
		if !result.IsError {
			t.Errorf("expected error for invalid URL %q", badURL)
		}
		if !strings.Contains(toolText(result), "http or https") {
			t.Errorf("expected helpful error message for URL %q, got: %s", badURL, toolText(result))
		}
	}
}

// ---- select_site ----

func TestSelectSiteTool(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"projects":[],"meta":{"page":1,"pages":1,"per_page":25,"count":0}}`)
	}))
	defer mock.Close()

	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	globalCfgPath := filepath.Join(dir, "config.yaml")
	t.Setenv("STQRY_CONFIG_HOME", dir)
	globalCfg := &config.GlobalConfig{
		Sites: map[string]*config.Site{
			"mysite": {Token: "tok-abc", APIURL: mock.URL},
		},
	}
	if err := config.SaveGlobalConfig(globalCfg, globalCfgPath); err != nil {
		t.Fatal(err)
	}

	s := stqrymcp.NewServer("")
	result := callTool(s, "select_site", `{"site_name":"mysite"}`)
	if result == nil || result.IsError {
		t.Fatalf("expected success, got error: %s", toolText(result))
	}

	// Session should now be set — list_projects should work.
	result = callTool(s, "list_projects", `{}`)
	if result == nil || result.IsError {
		t.Fatalf("list_projects failed after select_site: %s", toolText(result))
	}

	// No stqry.yaml should have been written to disk.
	if _, err := os.Stat(filepath.Join(dir, "stqry.yaml")); !os.IsNotExist(err) {
		t.Error("select_site should not write stqry.yaml to disk")
	}
}

func TestSelectSiteNotFound(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	t.Setenv("STQRY_CONFIG_HOME", dir)
	// No global config — site won't exist.

	s := stqrymcp.NewServer("")
	result := callTool(s, "select_site", `{"site_name":"unknown"}`)
	if result == nil {
		t.Fatal("expected a result")
	}
	if !result.IsError {
		t.Fatal("expected error for unknown site")
	}
}

func TestSelectSiteMissingSiteName(t *testing.T) {
	s := stqrymcp.NewServer("")
	result := callTool(s, "select_site", `{"site_name":""}`)
	if result == nil {
		t.Fatal("expected a result")
	}
	if !result.IsError {
		t.Fatal("expected error when site_name is empty")
	}
}

// ---- ResolveClient ----

func TestResolveClientFromDirConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "mytoken", APIURL: "https://api.example.com"}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	client, err := stqrymcp.ResolveClient("", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Token != "mytoken" {
		t.Errorf("expected mytoken, got %s", client.Token)
	}
}

func TestResolveClientFromSession(t *testing.T) {
	sess := stqrymcp.NewSession()
	sess.Set(&config.Site{Token: "session-tok", APIURL: "https://api.example.com"})

	client, err := stqrymcp.ResolveClient("", sess)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Token != "session-tok" {
		t.Errorf("expected session-tok, got %s", client.Token)
	}
}

func TestResolveClientNoConfigError(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)
	t.Setenv("STQRY_SITE", "")
	t.Setenv("STQRY_CONFIG_HOME", dir)

	_, err := stqrymcp.ResolveClient("", nil)
	if err == nil {
		t.Fatal("expected error when nothing configured")
	}
	if !strings.Contains(err.Error(), "connect(") {
		t.Errorf("expected helpful error mentioning connect(), got: %v", err)
	}
}

func TestResolveClientSessionPriorityOverDisk(t *testing.T) {
	// Write a stqry.yaml with a different token.
	dir := t.TempDir()
	diskCfg := &config.DirectoryConfig{Token: "disk-tok", APIURL: "https://disk.example.com"}
	if err := config.SaveDirectoryConfig(dir, diskCfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	// Session holds a different token and URL.
	sess := stqrymcp.NewSession()
	sess.Set(&config.Site{Token: "session-tok", APIURL: "https://session.example.com"})

	client, err := stqrymcp.ResolveClient("", sess)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Session must take priority over the disk config.
	if client.Token != "session-tok" {
		t.Errorf("expected session-tok to win over disk-tok, got %s", client.Token)
	}
	if client.BaseURL != "https://session.example.com" {
		t.Errorf("expected session APIURL to win over disk APIURL, got %s", client.BaseURL)
	}
}

// ---- list_projects ----

func TestListProjectsHappyPath(t *testing.T) {
	// Start a mock API server that returns a well-formed projects response.
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/projects" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"projects":[{"id":"42","name":"Test Project"}],"meta":{"page":1,"pages":1,"per_page":25,"count":1}}`)
	}))
	defer mock.Close()

	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "tok", APIURL: mock.URL}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "list_projects", `{}`)
	if result == nil {
		t.Fatal("expected a result")
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", toolText(result))
	}
	if !strings.Contains(toolText(result), "Test Project") {
		t.Errorf("expected project name in response, got: %s", toolText(result))
	}
}

func TestListProjectsAPIError(t *testing.T) {
	// Mock server returns a 500 error.
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errors":[{"code":"server_error","message":"internal error"}]}`, http.StatusInternalServerError)
	}))
	defer mock.Close()

	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "tok", APIURL: mock.URL}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "list_projects", `{}`)
	if result == nil {
		t.Fatal("expected a result")
	}
	if !result.IsError {
		t.Fatal("expected an error result when API returns 500")
	}
}

// ---- connect ----

func TestConnectTool(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"projects":[],"meta":{"page":1,"pages":1,"per_page":25,"count":0}}`)
	}))
	defer mock.Close()

	// Start in empty dir — no stqry.yaml, no global config.
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)
	t.Setenv("STQRY_CONFIG_HOME", dir)

	s := stqrymcp.NewServer("")

	// Without connect, list_projects should fail.
	result := callTool(s, "list_projects", `{}`)
	if !result.IsError {
		t.Fatal("expected error before connect")
	}

	// After connect, list_projects should succeed.
	result = callTool(s, "connect", fmt.Sprintf(`{"token":"tok","api_url":%q}`, mock.URL))
	if result == nil || result.IsError {
		t.Fatalf("connect failed: %s", toolText(result))
	}

	result = callTool(s, "list_projects", `{}`)
	if result == nil || result.IsError {
		t.Fatalf("list_projects failed after connect: %s", toolText(result))
	}
}

func TestConnectToolInvalidURL(t *testing.T) {
	s := stqrymcp.NewServer("")
	result := callTool(s, "connect", `{"token":"tok","api_url":"not-a-url"}`)
	if result == nil || !result.IsError {
		t.Fatal("expected error for invalid URL")
	}
	if !strings.Contains(toolText(result), "http or https") {
		t.Errorf("expected helpful error, got: %s", toolText(result))
	}
}

func TestConfigureProjectSetsSession(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"projects":[],"meta":{"page":1,"pages":1,"per_page":25,"count":0}}`)
	}))
	defer mock.Close()

	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)
	t.Setenv("STQRY_CONFIG_HOME", dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "configure_project", fmt.Sprintf(`{"api_url":%q,"token":"tok123"}`, mock.URL))
	if result == nil || result.IsError {
		t.Fatalf("configure_project failed: %s", toolText(result))
	}

	// Session should be set — list_projects should work without reading stqry.yaml.
	result = callTool(s, "list_projects", `{}`)
	if result == nil || result.IsError {
		t.Fatalf("list_projects failed after configure_project: %s", toolText(result))
	}
}

func TestListProjectsPagination(t *testing.T) {
	var receivedPage, receivedPerPage string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPage = r.URL.Query().Get("page")
		receivedPerPage = r.URL.Query().Get("per_page")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"projects":[],"meta":{"page":2,"pages":3,"per_page":10,"count":25}}`)
	}))
	defer mock.Close()

	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "tok", APIURL: mock.URL}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "list_projects", `{"page":2,"per_page":10}`)
	if result == nil {
		t.Fatal("expected a result")
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", toolText(result))
	}
	if receivedPage != "2" {
		t.Errorf("expected page=2 forwarded to API, got %q", receivedPage)
	}
	if receivedPerPage != "10" {
		t.Errorf("expected per_page=10 forwarded to API, got %q", receivedPerPage)
	}
}

func TestCreateMediaInvalidType(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.mp4")
	if err := os.WriteFile(filePath, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	s := stqrymcp.NewServer("")
	result := callTool(s, "create_media", fmt.Sprintf(
		`{"file_path":%q,"type":"invalid_type"}`, filePath,
	))
	if result == nil {
		t.Fatal("expected a result")
	}
	if !result.IsError {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(toolText(result), "invalid type") {
		t.Errorf("expected helpful error mentioning invalid type, got: %s", toolText(result))
	}
}

// ---- create_media ----

func TestCreateMediaMissingFilePath(t *testing.T) {
	s := stqrymcp.NewServer("")
	result := callTool(s, "create_media", `{"file_path":"","type":"video"}`)
	if result == nil || !result.IsError {
		t.Fatal("expected error for missing file_path")
	}
}

func TestCreateMediaMissingType(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "f.mp4")
	if err := os.WriteFile(filePath, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	s := stqrymcp.NewServer("")
	result := callTool(s, "create_media", fmt.Sprintf(`{"file_path":%q,"type":""}`, filePath))
	if result == nil || !result.IsError {
		t.Fatal("expected error for missing type")
	}
}

func TestCreateMediaBadFilePath(t *testing.T) {
	mux := http.NewServeMux()
	var mock *httptest.Server

	// Presign responds successfully; the error should come from the file open.
	mux.HandleFunc("/api/public/uploaded_files/presigned", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":    mock.URL + "/upload",
			"fields": map[string]string{"key": "uploads/nonexistent.mp4"},
		})
	})

	mock = httptest.NewServer(mux)
	defer mock.Close()

	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "tok", APIURL: mock.URL}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "create_media", `{"file_path":"/nonexistent/file.mp4","type":"video"}`)
	if result == nil || !result.IsError {
		t.Fatal("expected error for non-existent file path")
	}
}

func TestCreateMediaHappyPath(t *testing.T) {
	// Build temp file to upload.
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.mp4")
	if err := os.WriteFile(filePath, []byte("fake video content"), 0600); err != nil {
		t.Fatal(err)
	}

	// The mock server URL is referenced in the presigned handler closure,
	// so we declare the variable first and assign after NewServer.
	mux := http.NewServeMux()
	var mock *httptest.Server

	mux.HandleFunc("/api/public/uploaded_files/presigned", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url":    mock.URL + "/upload",
			"fields": map[string]string{"key": "uploads/test.mp4"},
		})
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/public/uploaded_files/process_enqueue", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"job_id": "job-media"})
	})
	mux.HandleFunc("/api/public/uploaded_files/process_status/job-media", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":        "complete",
			"uploaded_file": map[string]interface{}{"id": "uf-123"},
		})
	})
	mux.HandleFunc("/api/public/media_items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media_item": map[string]interface{}{"id": "mi-456", "type": "video"},
		})
	})

	mock = httptest.NewServer(mux)
	defer mock.Close()

	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "tok", APIURL: mock.URL}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "create_media", fmt.Sprintf(
		`{"file_path":%q,"type":"video","name":"Test Video"}`, filePath,
	))
	if result == nil || result.IsError {
		t.Fatalf("create_media failed: %s", toolText(result))
	}
	if !strings.Contains(toolText(result), "mi-456") {
		t.Errorf("expected media item id in response, got: %s", toolText(result))
	}
}

func TestCreateMediaUploadError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "bad.mp4")
	if err := os.WriteFile(filePath, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	// Presigned returns an error (API returns 500).
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errors":[{"code":"server_error"}]}`, http.StatusInternalServerError)
	}))
	defer mock.Close()

	dir := t.TempDir()
	cfg := &config.DirectoryConfig{Token: "tok", APIURL: mock.URL}
	if err := config.SaveDirectoryConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(dir)

	s := stqrymcp.NewServer("")
	result := callTool(s, "create_media", fmt.Sprintf(`{"file_path":%q,"type":"video"}`, filePath))
	if result == nil || !result.IsError {
		t.Fatal("expected error when upload API returns 500")
	}
}
