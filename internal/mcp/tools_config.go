package mcp

import (
	"context"
	"fmt"
	"net/url"
	"os"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/config"
)

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
		mcpgo.NewTool("debug_env",
			mcpgo.WithDescription("Dump environment variables and working directory for debugging. Remove before release."),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			cwd, _ := os.Getwd()
			env := os.Environ()
			return jsonResult(map[string]interface{}{
				"cwd": cwd,
				"env": env,
			})
		},
	)

	s.AddTool(
		mcpgo.NewTool("connect",
			mcpgo.WithDescription("Connect to a STQRY site for this session using a token and API URL. "+
				"Read these from stqry.yaml in your project directory. "+
				"Credentials are held in memory and cleared when the MCP server restarts."),
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
			token := req.GetString("token", "")
			apiURL := req.GetString("api_url", "")
			if token == "" || apiURL == "" {
				return mcpgo.NewToolResultError("token and api_url are required"), nil
			}
			parsed, err := url.Parse(apiURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return mcpgo.NewToolResultError("api_url must be a valid http or https URL (e.g. https://api.stqry.com)"), nil
			}
			sess.Set(&config.Site{Token: token, APIURL: apiURL})
			return jsonResult(map[string]interface{}{
				"ok":      true,
				"message": "connected to " + parsed.Host,
			})
		},
	)

	s.AddTool(
		mcpgo.NewTool("configure_project",
			mcpgo.WithDescription("Write stqry.yaml in the current directory with API credentials. Use this to configure a project to connect to STQRY."),
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
			apiURL := req.GetString("api_url", "")
			token := req.GetString("token", "")
			if apiURL == "" || token == "" {
				return mcpgo.NewToolResultError("api_url and token are required"), nil
			}
			parsed, err := url.Parse(apiURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return mcpgo.NewToolResultError("api_url must be a valid http or https URL (e.g. https://api.stqry.com)"), nil
			}
			if err := WriteProjectConfig(apiURL, token); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("writing config: %v", err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true,"message":"stqry.yaml written successfully"}`), nil
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
			siteName := req.GetString("site_name", "")
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
			if err := WriteProjectConfig(site.APIURL, site.Token); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("writing config: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"ok":      true,
				"message": fmt.Sprintf("switched to site %s", siteName),
			})
		},
	)
}
