package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/api"
)

// jsonResourceContents marshals v and returns it as a single-element
// []mcp.ResourceContents slice, or an error on marshalling failure.
func jsonResourceContents(uri string, v interface{}) ([]mcpgo.ResourceContents, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("serializing resource: %w", err)
	}
	return []mcpgo.ResourceContents{
		mcpgo.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func registerResources(s *server.MCPServer, flagSite string) {
	// stqry://projects — list of all projects
	s.AddResource(
		mcpgo.NewResource(
			"stqry://projects",
			"Projects",
			mcpgo.WithResourceDescription("All STQRY projects for the configured site."),
			mcpgo.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcpgo.ReadResourceRequest) ([]mcpgo.ResourceContents, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return nil, fmt.Errorf("resolving client: %w", err)
			}
			projects, meta, err := api.ListProjects(client, nil)
			if err != nil {
				return nil, fmt.Errorf("listing projects: %w", err)
			}
			return jsonResourceContents(req.Params.URI, map[string]interface{}{
				"projects": projects,
				"meta":     meta,
			})
		},
	)

	// stqry://collections — list of all collections
	s.AddResource(
		mcpgo.NewResource(
			"stqry://collections",
			"Collections",
			mcpgo.WithResourceDescription("All STQRY collections for the configured site."),
			mcpgo.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcpgo.ReadResourceRequest) ([]mcpgo.ResourceContents, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return nil, fmt.Errorf("resolving client: %w", err)
			}
			collections, meta, err := api.ListCollections(client, nil)
			if err != nil {
				return nil, fmt.Errorf("listing collections: %w", err)
			}
			return jsonResourceContents(req.Params.URI, map[string]interface{}{
				"collections": collections,
				"meta":        meta,
			})
		},
	)

	// stqry://screens — list of all screens
	s.AddResource(
		mcpgo.NewResource(
			"stqry://screens",
			"Screens",
			mcpgo.WithResourceDescription("All STQRY screens for the configured site."),
			mcpgo.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcpgo.ReadResourceRequest) ([]mcpgo.ResourceContents, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return nil, fmt.Errorf("resolving client: %w", err)
			}
			screens, meta, err := api.ListScreens(client, nil)
			if err != nil {
				return nil, fmt.Errorf("listing screens: %w", err)
			}
			return jsonResourceContents(req.Params.URI, map[string]interface{}{
				"screens": screens,
				"meta":    meta,
			})
		},
	)

	// stqry://media — list of all media items
	s.AddResource(
		mcpgo.NewResource(
			"stqry://media",
			"Media Items",
			mcpgo.WithResourceDescription("All STQRY media items for the configured site."),
			mcpgo.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcpgo.ReadResourceRequest) ([]mcpgo.ResourceContents, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return nil, fmt.Errorf("resolving client: %w", err)
			}
			items, meta, err := api.ListMediaItems(client, nil)
			if err != nil {
				return nil, fmt.Errorf("listing media items: %w", err)
			}
			return jsonResourceContents(req.Params.URI, map[string]interface{}{
				"media_items": items,
				"meta":        meta,
			})
		},
	)

	// stqry://codes — list of all codes
	s.AddResource(
		mcpgo.NewResource(
			"stqry://codes",
			"Codes",
			mcpgo.WithResourceDescription("All STQRY codes for the configured site."),
			mcpgo.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcpgo.ReadResourceRequest) ([]mcpgo.ResourceContents, error) {
			client, err := ResolveClient(flagSite)
			if err != nil {
				return nil, fmt.Errorf("resolving client: %w", err)
			}
			codes, meta, err := api.ListCodes(client, nil)
			if err != nil {
				return nil, fmt.Errorf("listing codes: %w", err)
			}
			return jsonResourceContents(req.Params.URI, map[string]interface{}{
				"codes": codes,
				"meta":  meta,
			})
		},
	)

	// stqry://collections/{id} — a single collection with its items
	s.AddResourceTemplate(
		mcpgo.NewResourceTemplate(
			"stqry://collections/{id}",
			"Collection",
			mcpgo.WithTemplateDescription("A single STQRY collection and its items, identified by ID."),
			mcpgo.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcpgo.ReadResourceRequest) ([]mcpgo.ResourceContents, error) {
			idVals, _ := req.Params.Arguments["id"].([]string)
			if len(idVals) == 0 || idVals[0] == "" {
				return nil, fmt.Errorf("collection ID is required")
			}
			id := idVals[0]
			client, err := ResolveClient(flagSite)
			if err != nil {
				return nil, fmt.Errorf("resolving client: %w", err)
			}
			collection, err := api.GetCollection(client, id)
			if err != nil {
				return nil, fmt.Errorf("getting collection: %w", err)
			}
			items, _, err := api.ListCollectionItems(client, id, nil)
			if err != nil {
				return nil, fmt.Errorf("listing collection items: %w", err)
			}
			return jsonResourceContents(req.Params.URI, map[string]interface{}{
				"collection":       collection,
				"collection_items": items,
			})
		},
	)

	// stqry://screens/{id} — a single screen with its story sections
	s.AddResourceTemplate(
		mcpgo.NewResourceTemplate(
			"stqry://screens/{id}",
			"Screen",
			mcpgo.WithTemplateDescription("A single STQRY screen and its story sections, identified by ID."),
			mcpgo.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcpgo.ReadResourceRequest) ([]mcpgo.ResourceContents, error) {
			idVals, _ := req.Params.Arguments["id"].([]string)
			if len(idVals) == 0 || idVals[0] == "" {
				return nil, fmt.Errorf("screen ID is required")
			}
			id := idVals[0]
			client, err := ResolveClient(flagSite)
			if err != nil {
				return nil, fmt.Errorf("resolving client: %w", err)
			}
			screen, err := api.GetScreen(client, id)
			if err != nil {
				return nil, fmt.Errorf("getting screen: %w", err)
			}
			sections, _, err := api.ListStorySections(client, id, nil)
			if err != nil {
				return nil, fmt.Errorf("listing story sections: %w", err)
			}
			return jsonResourceContents(req.Params.URI, map[string]interface{}{
				"screen":         screen,
				"story_sections": sections,
			})
		},
	)
}
