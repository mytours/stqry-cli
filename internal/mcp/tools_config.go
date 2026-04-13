package mcp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mytours/stqry-cli/internal/config"
)

// localConfigExists reports whether stqry.yaml or stqry.yml exists in the CWD.
// Intentionally checks only the current directory, not parent directories — the
// save suggestion is about creating a config here, not whether one governs the session.
func localConfigExists() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	for _, name := range []string{"stqry.yaml", "stqry.yml"} {
		if _, err := os.Stat(filepath.Join(cwd, name)); err == nil {
			return true
		}
	}
	return false
}

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
		mcpgo.NewTool("connect",
			mcpgo.WithDescription("Store site credentials in this session. "+
				"Credentials are held in memory and cleared when the MCP server restarts. "+
				"If you have a stqry.yaml file in the project directory, pass the token and api_url from it here."),
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
			token := strings.TrimSpace(req.GetString("token", ""))
			apiURL := strings.TrimSpace(req.GetString("api_url", ""))
			if token == "" || apiURL == "" {
				return mcpgo.NewToolResultError("token and api_url are required"), nil
			}
			parsed, err := url.Parse(apiURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return mcpgo.NewToolResultError("api_url must be a valid http or https URL (e.g. https://api.stqry.com)"), nil
			}
			sess.Set(&config.Site{Token: token, APIURL: apiURL})
			resp := map[string]interface{}{
				"ok":      true,
				"message": "connected to " + parsed.Host,
			}
			if !localConfigExists() {
				resp["save_suggested"] = true
				resp["save_message"] = "No stqry.yaml found in the current directory. " +
					"To save inline credentials locally, call configure_project(api_url, token). " +
					"To save as a named global site, call add_global_site(name, api_url, token) " +
					"then configure_project(site_name) to reference it here."
			}
			return jsonResult(resp)
		},
	)

	s.AddTool(
		mcpgo.NewTool("configure_project",
			mcpgo.WithDescription("Configure a STQRY project by storing credentials in this session. "+
				"Pass site_name to write a named reference (site: <name>) to stqry.yaml, or pass api_url+token "+
				"to write inline credentials. Also stores credentials in the session for immediate use."),
			mcpgo.WithString("site_name",
				mcpgo.Description("Named site from global config to reference in stqry.yaml (alternative to api_url+token)"),
			),
			mcpgo.WithString("api_url",
				mcpgo.Description("The STQRY API URL, e.g. https://api.stqry.com or https://api-us.stqry.com"),
			),
			mcpgo.WithString("token",
				mcpgo.Description("The STQRY API token for this site"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			siteName := strings.TrimSpace(req.GetString("site_name", ""))
			apiURL := strings.TrimSpace(req.GetString("api_url", ""))
			token := strings.TrimSpace(req.GetString("token", ""))

			if siteName != "" {
				// Named-reference mode: write site: <name> to stqry.yaml.
				globalCfg, err := config.LoadGlobalConfig(config.DefaultGlobalConfigPath())
				if err != nil {
					return mcpgo.NewToolResultError(fmt.Sprintf("loading global config: %v", err)), nil
				}
				if _, ok := globalCfg.Sites[siteName]; !ok {
					return mcpgo.NewToolResultError(fmt.Sprintf(
						"site %q not found in global config. Add it first with add_global_site or `stqry config add-site`",
						siteName,
					)), nil
				}
				cwd, err := os.Getwd()
				if err != nil {
					return mcpgo.NewToolResultError(fmt.Sprintf("getting working directory: %v", err)), nil
				}
				if err := config.SaveDirectoryConfig(cwd, &config.DirectoryConfig{Site: siteName}); err != nil {
					return mcpgo.NewToolResultError(fmt.Sprintf("writing stqry.yaml: %v", err)), nil
				}
				sess.Set(&config.Site{Token: globalCfg.Sites[siteName].Token, APIURL: globalCfg.Sites[siteName].APIURL})
				return jsonResult(map[string]interface{}{
					"ok":      true,
					"message": fmt.Sprintf("stqry.yaml written with site reference: %s", siteName),
				})
			}

			// Inline credentials mode (existing behaviour).
			if apiURL == "" || token == "" {
				return mcpgo.NewToolResultError("provide either site_name, or both api_url and token"), nil
			}
			parsed, err := url.Parse(apiURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return mcpgo.NewToolResultError("api_url must be a valid http or https URL (e.g. https://api.stqry.com)"), nil
			}
			site := &config.Site{Token: token, APIURL: apiURL}
			sess.Set(site)

			writeErr := WriteProjectConfig(apiURL, token)
			if writeErr != nil {
				return jsonResult(map[string]interface{}{
					"ok":      true,
					"message": fmt.Sprintf("connected (note: could not write stqry.yaml: %v)", writeErr),
				})
			}
			return jsonResult(map[string]interface{}{
				"ok":      true,
				"message": "stqry.yaml written successfully",
			})
		},
	)

	s.AddTool(
		mcpgo.NewTool("add_global_site",
			mcpgo.WithDescription("Add a named site to the global STQRY config (~/.config/stqry/config.yaml). "+
				"Use this to save credentials permanently so they can be referenced by name across projects. "+
				"Returns an error if a site with that name already exists."),
			mcpgo.WithString("name",
				mcpgo.Required(),
				mcpgo.Description("Name for this site (e.g. 'my-museum')"),
			),
			mcpgo.WithString("api_url",
				mcpgo.Required(),
				mcpgo.Description("The STQRY API URL, e.g. https://api.stqry.com"),
			),
			mcpgo.WithString("token",
				mcpgo.Required(),
				mcpgo.Description("The STQRY API token"),
			),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			name := strings.TrimSpace(req.GetString("name", ""))
			apiURL := strings.TrimSpace(req.GetString("api_url", ""))
			token := strings.TrimSpace(req.GetString("token", ""))
			if name == "" || apiURL == "" || token == "" {
				return mcpgo.NewToolResultError("name, api_url, and token are required"), nil
			}
			parsed, err := url.Parse(apiURL)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return mcpgo.NewToolResultError("api_url must be a valid http or https URL"), nil
			}
			globalCfg, err := config.LoadGlobalConfig(config.DefaultGlobalConfigPath())
			if err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("loading global config: %v", err)), nil
			}
			if _, exists := globalCfg.Sites[name]; exists {
				return mcpgo.NewToolResultError(fmt.Sprintf(
					"site %q already exists. Choose a different name or update it with `stqry config edit-site %s`",
					name, name,
				)), nil
			}
			globalCfg.Sites[name] = &config.Site{Token: token, APIURL: apiURL}
			globalCfgPath := config.DefaultGlobalConfigPath()
			if err := config.SaveGlobalConfig(globalCfg, globalCfgPath); err != nil {
				return mcpgo.NewToolResultError(fmt.Sprintf("saving global config: %v", err)), nil
			}
			return jsonResult(map[string]interface{}{
				"ok":      true,
				"message": fmt.Sprintf("site %q added to global config (%s)", name, globalCfgPath),
			})
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
			siteName := strings.TrimSpace(req.GetString("site_name", ""))
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
			sess.Set(&config.Site{Token: site.Token, APIURL: site.APIURL})
			resp := map[string]interface{}{
				"ok":      true,
				"message": fmt.Sprintf("switched to site %s", siteName),
			}
			if !localConfigExists() {
				resp["save_suggested"] = true
				resp["save_message"] = fmt.Sprintf(
					"No stqry.yaml found in the current directory. "+
						"Call configure_project(site_name: %q) to write a reference to this site here.", siteName)
			}
			return jsonResult(resp)
		},
	)
}
