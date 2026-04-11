package mcp

import (
	"fmt"
	"os"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/config"
)

// ResolveClient resolves the STQRY API client using the standard resolution order:
// --site flag → STQRY_SITE env var → stqry.yaml in CWD → global config.
func ResolveClient(flagSite string) (*api.Client, error) {
	globalCfg, err := config.LoadGlobalConfig(config.DefaultGlobalConfigPath())
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}
	dirCfg, err := config.FindDirectoryConfig(cwd)
	if err != nil {
		return nil, fmt.Errorf("finding directory config: %w", err)
	}
	site, err := config.ResolveSite(globalCfg, flagSite, dirCfg)
	if err != nil {
		return nil, err
	}
	return api.NewClient(site.APIURL, site.Token), nil
}

// jsonResult marshals v to a JSON tool result, or an error result on failure.
func jsonResult(v interface{}) (*mcpgo.CallToolResult, error) {
	data, err := marshalJSON(v)
	if err != nil {
		return mcpgo.NewToolResultError(fmt.Sprintf("serializing result: %v", err)), nil
	}
	return mcpgo.NewToolResultText(string(data)), nil
}

// NewServer creates the MCP server with all tools and resources registered.
func NewServer(flagSite string) *server.MCPServer {
	s := server.NewMCPServer("STQRY", "1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, false),
	)
	registerConfigTools(s, flagSite)
	registerProjectTools(s, flagSite)
	registerCollectionTools(s, flagSite)
	registerScreenTools(s, flagSite)
	registerMediaTools(s, flagSite)
	registerCodeTools(s, flagSite)
	registerResources(s, flagSite)
	return s
}

// Serve starts the MCP server on stdio. Blocks until client disconnects.
func Serve(flagSite string) error {
	return server.ServeStdio(NewServer(flagSite))
}
