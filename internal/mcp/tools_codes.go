package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
)

func registerCodeTools(s *server.MCPServer, flagSite string, sess *Session) {
	// list_codes: returns all codes for the configured site
	s.AddTool(
		mcpgo.NewTool("list_codes",
			mcpgo.WithDescription("List all codes for the configured STQRY site."),
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
			codes, meta, err := api.ListCodes(client, paginationQuery(req.GetInt("page", 0), req.GetInt("per_page", 0)))
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing codes: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"codes": codes,
				"meta":  meta,
			})
		},
	)

	// get_code: returns a single code by ID
	s.AddTool(
		mcpgo.NewTool("get_code",
			mcpgo.WithDescription("Get a single STQRY code by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The code ID"),
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
			code, err := api.GetCode(client, id)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("getting code: %v", err)), nil
			}
			return jsonResult(code)
		},
	)

	// create_code: creates a new code
	s.AddTool(
		mcpgo.NewTool("create_code",
			mcpgo.WithDescription("Create a new STQRY code."),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of code fields to set"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			args := req.GetArguments()
			fields, ok := args["fields"].(map[string]interface{})
			if !ok || fields == nil {
				return mcpgo.NewToolResultError("fields is required and must be an object"), nil
			}
			client, err := ResolveClient(flagSite, sess)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			code, err := api.CreateCode(client, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("creating code: %v", err)), nil
			}
			return jsonResult(code)
		},
	)

	// update_code: updates an existing code by ID
	s.AddTool(
		mcpgo.NewTool("update_code",
			mcpgo.WithDescription("Update an existing STQRY code by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The code ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of code fields to update"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			args := req.GetArguments()
			fields, ok := args["fields"].(map[string]interface{})
			if !ok || fields == nil {
				return mcpgo.NewToolResultError("fields is required and must be an object"), nil
			}
			client, err := ResolveClient(flagSite, sess)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			code, err := api.UpdateCode(client, id, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("updating code: %v", err)), nil
			}
			return jsonResult(code)
		},
	)

	// delete_code: deletes a code by ID
	s.AddTool(
		mcpgo.NewTool("delete_code",
			mcpgo.WithDescription("Delete a STQRY code by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The code ID"),
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
			if err := api.DeleteCode(client, id); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("deleting code: %v", err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true}`), nil
		},
	)
}
