package mcp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/config"
)

// localConfigExists reports whether stqry.yaml or stqry.yml exists in the CWD.
func localConfigExists() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	for _, name := range []string{"stqry.yaml", "stqry.yml"} {
		if _, err := os.Stat(filepath.Join(cwd, name)); err == nil {
			return true
		}
	}
	return false
}

// WriteProjectConfig writes a stqry.yaml with inline credentials to the CWD.
// Exported for testing.
func WriteProjectConfig(apiURL, token string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	cfg := &config.DirectoryConfig{
		Token:  token,
		APIURL: apiURL,
	}
	return config.SaveDirectoryConfig(cwd, cfg)
}

func registerConfigTools(s *server.MCPServer, flagSite string, sess *Session) {
	s.AddTool(
		mcpgo.NewTool("connect",
			mcpgo.WithDescription("Store site credentials in this session. "+
				"Credentials are held in memory and cleared when the MCP server restarts. "+
				"If you have a stqry.yaml file in the project directory, pass the token and api_url from it here."),
			mcpgo.WithString("token",
				mcpgo.Required(),
				mcpgo.Description("The STQRY API token"),
			),
			mcpgo.WithString("api_url",
				mcpgo.Required(),
				mcpgo.Description("The STQRY API URL, e.g. https://api.stqry.com or https://api-ca.stqry.com"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			token := strings.TrimSpace(req.GetString("token", ""))
			apiURL := strings.TrimSpace(req.GetString("api_url", ""))
			if token == "" || apiURL == "" {
				return mcpgo.NewToolResultError("token and api_url are required"), nil
			}
			parsed, err := url.Parse(apiURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return mcpgo.NewToolResultError("api_url must be a valid http or https URL (e.g. https://api.stqry.com)"), nil
			}
			sess.Set(&config.Site{Token: token, APIURL: apiURL})
			resp := map[string]interface{}{
				"ok":      true,
				"message": "connected to " + parsed.Host,
			}
			if !localConfigExists() {
				resp["save_suggested"] = true
				resp["save_message"] = "No stqry.yaml found in the current directory. Would you like to save these credentials? Options: locally (stqry.yaml here), globally (named site in ~/.config/stqry/config.yaml), or both."
			}
			return jsonResult(resp)
		},
	)

	s.AddTool(
		mcpgo.NewTool("configure_project",
			mcpgo.WithDescription("Configure a STQRY project by storing credentials in this session. "+
				"Also attempts to write stqry.yaml to the current directory for future use "+
				"(this may fail in read-only environments and is not fatal)."),
			mcpgo.WithString("api_url",
				mcpgo.Required(),
				mcpgo.Description("The STQRY API URL, e.g. https://api.stqry.com or https://api-us.stqry.com"),
			),
			mcpgo.WithString("token",
				mcpgo.Required(),
				mcpgo.Description("The STQRY API token for this site"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			apiURL := strings.TrimSpace(req.GetString("api_url", ""))
			token := strings.TrimSpace(req.GetString("token", ""))
			if apiURL == "" || token == "" {
				return mcpgo.NewToolResultError("api_url and token are required"), nil
			}
			parsed, err := url.Parse(apiURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return mcpgo.NewToolResultError("api_url must be a valid http or https URL (e.g. https://api.stqry.com)"), nil
			}
			site := &config.Site{Token: token, APIURL: apiURL}
			sess.Set(site)

			// Best-effort disk write — not fatal if CWD is read-only.
			writeErr := WriteProjectConfig(apiURL, token)
			if writeErr != nil {
				return jsonResult(map[string]interface{}{
					"ok":      true,
					"message": fmt.Sprintf("connected (note: could not write stqry.yaml: %v)", writeErr),
				})
			}
			return jsonResult(map[string]interface{}{
				"ok":      true,
				"message": "stqry.yaml written successfully",
			})
		},
	)

	s.AddTool(
		mcpgo.NewTool("select_site",
			mcpgo.WithDescription("Switch to a named site from global config (~/.config/stqry/config.yaml). Use this when the user says which site they want to work on."),
			mcpgo.WithString("site_name",
				mcpgo.Required(),
				mcpgo.Description("The site name as configured via `stqry config add-site`"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			siteName := strings.TrimSpace(req.GetString("site_name", ""))
			if siteName == "" {
				return mcpgo.NewToolResultError("site_name is required"), nil
			}
			globalCfg, err := config.LoadGlobalConfig(config.DefaultGlobalConfigPath())
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("loading global config: %v", err)), nil
			}
			site, ok := globalCfg.Sites[siteName]
			if !ok {
				return mcpgo.NewToolResultError(fmt.Sprintf("site %q not found. Run `stqry config add-site --name=%s --token=<token> --api-url=<url>` to add it", siteName, siteName)), nil
			}
			sess.Set(&config.Site{Token: site.Token, APIURL: site.APIURL})
			resp := map[string]interface{}{
				"ok":      true,
				"message": fmt.Sprintf("switched to site %s", siteName),
			}
			if !localConfigExists() {
				resp["save_suggested"] = true
				resp["save_message"] = fmt.Sprintf(
					"No stqry.yaml found in the current directory. Would you like to save a reference to this site locally? "+
						"I can write 'site: %s' to stqry.yaml in the current folder.", siteName)
			}
			return jsonResult(resp)
		},
	)
}
