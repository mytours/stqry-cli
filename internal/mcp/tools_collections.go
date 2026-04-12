package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
)

func registerCollectionTools(s *server.MCPServer, flagSite string) {
	// list_collections: returns all collections for the configured site
	s.AddTool(
		mcpgo.NewTool("list_collections",
			mcpgo.WithDescription("List all collections for the configured STQRY site."),
			mcpgo.WithNumber("page",
				mcpgo.Description("Page number (1-based, default: 1)"),
			),
			mcpgo.WithNumber("per_page",
				mcpgo.Description("Items per page (default: 25)"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			items, meta, err := api.ListCollections(client, paginationQuery(req.GetInt("page", 0), req.GetInt("per_page", 0)))
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing collections: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"collections": items,
				"meta":        meta,
			})
		},
	)

	// get_collection: returns a single collection by ID
	s.AddTool(
		mcpgo.NewTool("get_collection",
			mcpgo.WithDescription("Get a single STQRY collection by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The collection ID"),
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
			collection, err := api.GetCollection(client, id)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("getting collection: %v", err)), nil
			}
			return jsonResult(collection)
		},
	)

	// create_collection: creates a new collection with the given fields
	s.AddTool(
		mcpgo.NewTool("create_collection",
			mcpgo.WithDescription("Create a new STQRY collection."),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of collection fields to set"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			args := req.GetArguments()
			fields, ok := args["fields"].(map[string]interface{})
			if !ok || fields == nil {
				return mcpgo.NewToolResultError("fields is required and must be an object"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			collection, err := api.CreateCollection(client, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("creating collection: %v", err)), nil
			}
			return jsonResult(collection)
		},
	)

	// update_collection: updates an existing collection by ID
	s.AddTool(
		mcpgo.NewTool("update_collection",
			mcpgo.WithDescription("Update an existing STQRY collection by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The collection ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of collection fields to update"),
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
			collection, err := api.UpdateCollection(client, id, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("updating collection: %v", err)), nil
			}
			return jsonResult(collection)
		},
	)

	// delete_collection: deletes a collection by ID
	s.AddTool(
		mcpgo.NewTool("delete_collection",
			mcpgo.WithDescription("Delete a STQRY collection by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The collection ID"),
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
			if err := api.DeleteCollection(client, id); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("deleting collection: %v", err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true}`), nil
		},
	)

	// list_collection_items: returns all items for a collection
	s.AddTool(
		mcpgo.NewTool("list_collection_items",
			mcpgo.WithDescription("List all items in a STQRY collection."),
			mcpgo.WithString("collection_id",
				mcpgo.Required(),
				mcpgo.Description("The collection ID"),
			),
			mcpgo.WithNumber("page",
				mcpgo.Description("Page number (1-based, default: 1)"),
			),
			mcpgo.WithNumber("per_page",
				mcpgo.Description("Items per page (default: 25)"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			collectionID := req.GetString("collection_id", "")
			if collectionID == "" {
				return mcpgo.NewToolResultError("collection_id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			items, meta, err := api.ListCollectionItems(client, collectionID, paginationQuery(req.GetInt("page", 0), req.GetInt("per_page", 0)))
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing collection items: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"items": items,
				"meta":  meta,
			})
		},
	)

	// create_collection_item: creates a new item in a collection
	s.AddTool(
		mcpgo.NewTool("create_collection_item",
			mcpgo.WithDescription("Create a new item in a STQRY collection."),
			mcpgo.WithString("collection_id",
				mcpgo.Required(),
				mcpgo.Description("The collection ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of item fields to set"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			collectionID := req.GetString("collection_id", "")
			if collectionID == "" {
				return mcpgo.NewToolResultError("collection_id is required"), nil
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
			item, err := api.CreateCollectionItem(client, collectionID, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("creating collection item: %v", err)), nil
			}
			return jsonResult(item)
		},
	)

	// update_collection_item: updates an existing collection item
	s.AddTool(
		mcpgo.NewTool("update_collection_item",
			mcpgo.WithDescription("Update an existing item in a STQRY collection."),
			mcpgo.WithString("collection_id",
				mcpgo.Required(),
				mcpgo.Description("The collection ID"),
			),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The item ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of item fields to update"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			collectionID := req.GetString("collection_id", "")
			if collectionID == "" {
				return mcpgo.NewToolResultError("collection_id is required"), nil
			}
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
			item, err := api.UpdateCollectionItem(client, collectionID, id, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("updating collection item: %v", err)), nil
			}
			return jsonResult(item)
		},
	)

	// delete_collection_item: deletes an item from a collection
	s.AddTool(
		mcpgo.NewTool("delete_collection_item",
			mcpgo.WithDescription("Delete an item from a STQRY collection."),
			mcpgo.WithString("collection_id",
				mcpgo.Required(),
				mcpgo.Description("The collection ID"),
			),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The item ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			collectionID := req.GetString("collection_id", "")
			if collectionID == "" {
				return mcpgo.NewToolResultError("collection_id is required"), nil
			}
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			if err := api.DeleteCollectionItem(client, collectionID, id); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("deleting collection item: %v", err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true}`), nil
		},
	)
}
