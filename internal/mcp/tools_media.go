package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
)

func registerMediaTools(s *server.MCPServer, flagSite string) {
	// list_media: returns all media items for the configured site
	s.AddTool(
		mcpgo.NewTool("list_media",
			mcpgo.WithDescription("List all media items for the configured STQRY site."),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			items, meta, err := api.ListMediaItems(client, nil)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing media items: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"media_items": items,
				"meta":        meta,
			})
		},
	)

	// get_media: returns a single media item by ID
	s.AddTool(
		mcpgo.NewTool("get_media",
			mcpgo.WithDescription("Get a single STQRY media item by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The media item ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			item, err := api.GetMediaItem(client, id)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("getting media item: %v", err)), nil
			}
			return jsonResult(item)
		},
	)

	// update_media: updates an existing media item by ID
	s.AddTool(
		mcpgo.NewTool("update_media",
			mcpgo.WithDescription("Update an existing STQRY media item by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The media item ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of media item fields to update"),
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
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			item, err := api.UpdateMediaItem(client, id, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("updating media item: %v", err)), nil
			}
			return jsonResult(item)
		},
	)

	// delete_media: deletes a media item by ID
	s.AddTool(
		mcpgo.NewTool("delete_media",
			mcpgo.WithDescription("Delete a STQRY media item by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The media item ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			if err := api.DeleteMediaItem(client, id, nil); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("deleting media item: %v", err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true}`), nil
		},
	)
}
