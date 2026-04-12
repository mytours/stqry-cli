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

func registerConfigTools(s *server.MCPServer, flagSite string) {
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
}
