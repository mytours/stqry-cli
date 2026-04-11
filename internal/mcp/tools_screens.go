package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
)

func registerScreenTools(s *server.MCPServer, flagSite string) {
	registerScreenCRUD(s, flagSite)
	registerSectionCRUD(s, flagSite)
	registerSubItemTools(s, flagSite, "list_badge_items", "create_badge_item", "update_badge_item", "delete_badge_item", "badge_items", "badge_item")
	registerSubItemTools(s, flagSite, "list_link_items", "create_link_item", "update_link_item", "delete_link_item", "link_items", "link_item")
	registerSubItemTools(s, flagSite, "list_section_media", "create_section_media", "update_section_media", "delete_section_media", "media_items", "media_item")
	registerSubItemTools(s, flagSite, "list_price_items", "create_price_item", "update_price_item", "delete_price_item", "price_items", "price_item")
	registerSubItemTools(s, flagSite, "list_social_items", "create_social_item", "update_social_item", "delete_social_item", "social_items", "social_item")
	registerSubItemTools(s, flagSite, "list_opening_times", "create_opening_time", "update_opening_time", "delete_opening_time", "opening_time_items", "opening_time_item")
}

func registerScreenCRUD(s *server.MCPServer, flagSite string) {
	// list_screens: returns all screens for the configured site
	s.AddTool(
		mcpgo.NewTool("list_screens",
			mcpgo.WithDescription("List all screens for the configured STQRY site."),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			items, meta, err := api.ListScreens(client, nil)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing screens: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"screens": items,
				"meta":    meta,
			})
		},
	)

	// get_screen: returns a single screen by ID
	s.AddTool(
		mcpgo.NewTool("get_screen",
			mcpgo.WithDescription("Get a single STQRY screen by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
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
			screen, err := api.GetScreen(client, id)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("getting screen: %v", err)), nil
			}
			return jsonResult(screen)
		},
	)

	// create_screen: creates a new screen with the given fields
	s.AddTool(
		mcpgo.NewTool("create_screen",
			mcpgo.WithDescription("Create a new STQRY screen."),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of screen fields to set"),
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
			screen, err := api.CreateScreen(client, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("creating screen: %v", err)), nil
			}
			return jsonResult(screen)
		},
	)

	// update_screen: updates an existing screen by ID
	s.AddTool(
		mcpgo.NewTool("update_screen",
			mcpgo.WithDescription("Update an existing STQRY screen by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of screen fields to update"),
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
			screen, err := api.UpdateScreen(client, id, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("updating screen: %v", err)), nil
			}
			return jsonResult(screen)
		},
	)

	// delete_screen: deletes a screen by ID
	s.AddTool(
		mcpgo.NewTool("delete_screen",
			mcpgo.WithDescription("Delete a STQRY screen by ID."),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
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
			if err := api.DeleteScreen(client, id); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("deleting screen: %v", err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true}`), nil
		},
	)
}

func registerSectionCRUD(s *server.MCPServer, flagSite string) {
	// list_sections: returns all story sections for a screen
	s.AddTool(
		mcpgo.NewTool("list_sections",
			mcpgo.WithDescription("List all story sections for a STQRY screen."),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			items, meta, err := api.ListStorySections(client, screenID, nil)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing sections: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"story_sections": items,
				"meta":           meta,
			})
		},
	)

	// get_section: returns a single story section by ID
	s.AddTool(
		mcpgo.NewTool("get_section",
			mcpgo.WithDescription("Get a single STQRY story section by ID."),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The section ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
			}
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			section, err := api.GetStorySection(client, screenID, id)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("getting section: %v", err)), nil
			}
			return jsonResult(section)
		},
	)

	// create_section: creates a new story section for a screen
	s.AddTool(
		mcpgo.NewTool("create_section",
			mcpgo.WithDescription("Create a new STQRY story section for a screen."),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of section fields to set"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
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
			section, err := api.CreateStorySection(client, screenID, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("creating section: %v", err)), nil
			}
			return jsonResult(section)
		},
	)

	// update_section: updates an existing story section
	s.AddTool(
		mcpgo.NewTool("update_section",
			mcpgo.WithDescription("Update an existing STQRY story section by ID."),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The section ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description("Arbitrary JSON object of section fields to update"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
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
			section, err := api.UpdateStorySection(client, screenID, id, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("updating section: %v", err)), nil
			}
			return jsonResult(section)
		},
	)

	// delete_section: deletes a story section by ID
	s.AddTool(
		mcpgo.NewTool("delete_section",
			mcpgo.WithDescription("Delete a STQRY story section by ID."),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description("The section ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
			}
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			if err := api.DeleteStorySection(client, screenID, id); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("deleting section: %v", err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true}`), nil
		},
	)
}

func registerSubItemTools(s *server.MCPServer, flagSite string, listTool, createTool, updateTool, deleteTool, apiPath, singularKey string) {
	// list sub-items for a section
	s.AddTool(
		mcpgo.NewTool(listTool,
			mcpgo.WithDescription(fmt.Sprintf("List all %s for a STQRY story section.", apiPath)),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithString("section_id",
				mcpgo.Required(),
				mcpgo.Description("The section ID"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
			}
			sectionID := req.GetString("section_id", "")
			if sectionID == "" {
				return mcpgo.NewToolResultError("section_id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			items, err := api.ListSectionSubItems(client, screenID, sectionID, apiPath)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("listing %s: %v", apiPath, err)), nil
			}
			return jsonResult(map[string]interface{}{
				apiPath: items,
			})
		},
	)

	// create a sub-item in a section
	s.AddTool(
		mcpgo.NewTool(createTool,
			mcpgo.WithDescription(fmt.Sprintf("Create a new %s in a STQRY story section.", singularKey)),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithString("section_id",
				mcpgo.Required(),
				mcpgo.Description("The section ID"),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description(fmt.Sprintf("Arbitrary JSON object of %s fields to set", singularKey)),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
			}
			sectionID := req.GetString("section_id", "")
			if sectionID == "" {
				return mcpgo.NewToolResultError("section_id is required"), nil
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
			item, err := api.CreateSectionSubItem(client, screenID, sectionID, apiPath, singularKey, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("creating %s: %v", singularKey, err)), nil
			}
			return jsonResult(item)
		},
	)

	// update a sub-item in a section
	s.AddTool(
		mcpgo.NewTool(updateTool,
			mcpgo.WithDescription(fmt.Sprintf("Update an existing %s in a STQRY story section.", singularKey)),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithString("section_id",
				mcpgo.Required(),
				mcpgo.Description("The section ID"),
			),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description(fmt.Sprintf("The %s ID", singularKey)),
			),
			mcpgo.WithObject("fields",
				mcpgo.Required(),
				mcpgo.Description(fmt.Sprintf("Arbitrary JSON object of %s fields to update", singularKey)),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
			}
			sectionID := req.GetString("section_id", "")
			if sectionID == "" {
				return mcpgo.NewToolResultError("section_id is required"), nil
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
			item, err := api.UpdateSectionSubItem(client, screenID, sectionID, apiPath, id, singularKey, fields)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("updating %s: %v", singularKey, err)), nil
			}
			return jsonResult(item)
		},
	)

	// delete a sub-item from a section
	s.AddTool(
		mcpgo.NewTool(deleteTool,
			mcpgo.WithDescription(fmt.Sprintf("Delete a %s from a STQRY story section.", singularKey)),
			mcpgo.WithString("screen_id",
				mcpgo.Required(),
				mcpgo.Description("The screen ID"),
			),
			mcpgo.WithString("section_id",
				mcpgo.Required(),
				mcpgo.Description("The section ID"),
			),
			mcpgo.WithString("id",
				mcpgo.Required(),
				mcpgo.Description(fmt.Sprintf("The %s ID", singularKey)),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			screenID := req.GetString("screen_id", "")
			if screenID == "" {
				return mcpgo.NewToolResultError("screen_id is required"), nil
			}
			sectionID := req.GetString("section_id", "")
			if sectionID == "" {
				return mcpgo.NewToolResultError("section_id is required"), nil
			}
			id := req.GetString("id", "")
			if id == "" {
				return mcpgo.NewToolResultError("id is required"), nil
			}
			client, err := ResolveClient(flagSite)
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("resolving client: %v", err)), nil
			}
			if err := api.DeleteSectionSubItem(client, screenID, sectionID, apiPath, id); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("deleting %s: %v", singularKey, err)), nil
			}
			return mcpgo.NewToolResultText(`{"ok":true}`), nil
		},
	)
}
