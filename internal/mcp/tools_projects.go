package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
)

func registerProjectTools(s *server.MCPServer, flagSite string, sess *Session) {
	// list_projects: returns all projects for the configured site
	s.AddTool(
		mcpgo.NewTool("list_projects",
			mcpgo.WithDescription("List all projects for the configured STQRY site."),
			mcpgo.WithNumber("page",
				mcpgo.Description("Page number (1-based, default: 1)"),
			),
			mcpgo.WithNumber("per_page",
				mcpgo.Description("Items per page (default: 25)"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			client, err := ResolveClient(flagSite, sess)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			projects, meta, err := api.ListProjects(client, paginationQuery(req.GetInt("page", 0), req.GetInt("per_page", 0)))
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing projects: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"projects": projects,
				"meta":     meta,
			})
		},
	)

	// get_project: returns a single project by ID
	s.AddTool(
		mcpgo.NewTool("get_project",
			mcpgo.WithDescription("Get a single STQRY project by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The project ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			client, err := ResolveClient(flagSite, sess)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			project, err := api.GetProject(client, id)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("getting project: %v", err)), nil
			}
			return jsonResult(project)
		},
	)
}
