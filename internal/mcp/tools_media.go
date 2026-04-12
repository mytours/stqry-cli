package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
)

func registerMediaTools(s *server.MCPServer, flagSite string, sess *Session) {
	// create_media: uploads a file and creates a new media item
	s.AddTool(
		mcpgo.NewTool("create_media",
			mcpgo.WithDescription("Upload a file and create a new STQRY media item."),
			mcpgo.WithString("file_path",
				mcpgo.Required(),
				mcpgo.Description("Absolute path to the file to upload"),
			),
			mcpgo.WithString("type",
				mcpgo.Required(),
				mcpgo.Description("Media item type: map, webpackage, animation, audio, image, video, webvideo, ar, data"),
			),
			mcpgo.WithString("name",
				mcpgo.Description("Name for the media item"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			filePath := req.GetString("file_path", "")
			if filePath == "" {
				return mcpgo.NewToolResultError("file_path is required"), nil
			}
			mediaType := req.GetString("type", "")
			if mediaType == "" {
				return mcpgo.NewToolResultError("type is required"), nil
			}
			name := req.GetString("name", "")

			client, err := ResolveClient(flagSite, sess)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}

			uploadedFile, err := api.UploadFile(client, filePath, "", nil, nil)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("uploading file: %v", err)), nil
			}

			uploadedFileID := ""
			if id, ok := uploadedFile["id"].(string); ok {
				uploadedFileID = id
			} else if id, ok := uploadedFile["id"].(float64); ok {
				uploadedFileID = fmt.Sprintf("%d", int(id))
			}
			if uploadedFileID == "" {
				return mcpgo.NewToolResultError("upload succeeded but returned no file ID"), nil
			}

			fields := map[string]interface{}{
				"type":                  mediaType,
				"file_uploaded_file_id": uploadedFileID,
			}
			if name != "" {
				fields["name"] = name
			}

			item, err := api.CreateMediaItem(client, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("creating media item: %v", err)), nil
			}
			return jsonResult(item)
		},
	)

	// list_media: returns all media items for the configured site
	s.AddTool(
		mcpgo.NewTool("list_media",
			mcpgo.WithDescription("List all media items for the configured STQRY site."),
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
			items, meta, err := api.ListMediaItems(client, paginationQuery(req.GetInt("page", 0), req.GetInt("per_page", 0)))
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
			client, err := ResolveClient(flagSite, sess)
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
			client, err := ResolveClient(flagSite, sess)
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
			client, err := ResolveClient(flagSite, sess)
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
