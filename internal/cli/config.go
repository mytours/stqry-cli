package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/mytours/stqry-cli/internal/config"
	"github.com/spf13/cobra"
)

var regionURLs = map[string]string{
	"us": "https://api-us.stqry.com",
	"ca": "https://api-ca.stqry.com",
	"eu": "https://api-eu.stqry.com",
	"sg": "https://api-sg.stqry.com",
	"au": "https://api-au.stqry.com",
}

func resolveAPIURL(region, apiURL string) (string, error) {
	if apiURL != "" {
		return apiURL, nil
	}
	if region != "" {
		url, ok := regionURLs[region]
		if !ok {
			return "", fmt.Errorf("unknown region %q. Valid regions: us, ca, eu, sg, au", region)
		}
		return url, nil
	}
	return "", fmt.Errorf("either --region or --api-url is required")
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Example: `  # Add a US site to global config
  stqry config add-site --name mysite --token abc123 --region us

  # See all configured sites
  stqry config list-sites`,
	}

	cmd.AddCommand(newConfigAddSiteCmd())
	cmd.AddCommand(newConfigRemoveSiteCmd())
	cmd.AddCommand(newConfigEditSiteCmd())
	cmd.AddCommand(newConfigListSitesCmd())
	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())

	return cmd
}

// config add-site --name=X --token=X --region=X [--api-url=X]
func newConfigAddSiteCmd() *cobra.Command {
	var name, token, region, apiURL string

	cmd := &cobra.Command{
		Use:   "add-site",
		Short: "Add a site to global config",
		Example: `  # Add a US-region site
  stqry config add-site --name mysite --token abc123 --region us

  # Add a site with a custom API URL (e.g. staging)
  stqry config add-site --name staging --token xyz --api-url https://api-staging.example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if token == "" {
				return fmt.Errorf("--token is required")
			}
			resolvedURL, err := resolveAPIURL(region, apiURL)
			if err != nil {
				return err
			}

			if _, exists := globalConfig.Sites[name]; exists {
				return fmt.Errorf("site %q already exists. Use `stqry config edit-site %s` to update it", name, name)
			}

			globalConfig.Sites[name] = &config.Site{
				Token:  token,
				APIURL: resolvedURL,
			}

			if err := config.SaveGlobalConfig(globalConfig, config.DefaultGlobalConfigPath()); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !flagQuiet && !flagJSON {
				fmt.Printf("Site %q added.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Site name (required)")
	cmd.Flags().StringVar(&token, "token", "", "API token (required)")
	cmd.Flags().StringVar(&region, "region", "", "Region (us, ca, eu, sg, au)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API base URL (overrides --region, e.g. for staging)")

	return cmd
}

// config remove-site <name>
func newConfigRemoveSiteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-site <name>",
		Short: "Remove a site from global config",
		Example: `  # Remove a site from global config
  stqry config remove-site mysite`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if _, exists := globalConfig.Sites[name]; !exists {
				return fmt.Errorf("site %q not found in config", name)
			}

			delete(globalConfig.Sites, name)

			if err := config.SaveGlobalConfig(globalConfig, config.DefaultGlobalConfigPath()); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !flagQuiet && !flagJSON {
				fmt.Printf("Site %q removed.\n", name)
			}
			return nil
		},
	}
}

// config edit-site <name> [--token=X] [--api-url=X]
func newConfigEditSiteCmd() *cobra.Command {
	var token, apiURL string

	cmd := &cobra.Command{
		Use:   "edit-site <name>",
		Short: "Update an existing site's fields",
		Example: `  # Update the API token for a site
  stqry config edit-site mysite --token newtoken123

  # Point a site at a different API URL
  stqry config edit-site mysite --api-url https://api-eu.stqry.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			site, exists := globalConfig.Sites[name]
			if !exists {
				return fmt.Errorf("site %q not found in config. Use `stqry config add-site` to create it", name)
			}

			if token != "" {
				site.Token = token
			}
			if apiURL != "" {
				site.APIURL = apiURL
			}

			if err := config.SaveGlobalConfig(globalConfig, config.DefaultGlobalConfigPath()); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			if !flagQuiet && !flagJSON {
				fmt.Printf("Site %q updated.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "New API token")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "New API base URL")

	return cmd
}

// config list-sites
func newConfigListSitesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-sites",
		Short: "List all configured sites",
		Example: `  # Show all configured sites
  stqry config list-sites`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(globalConfig.Sites) == 0 {
				if !flagQuiet && !flagJSON {
					fmt.Println("No sites configured. Use `stqry config add-site` to add one.")
				} else {
					printer.PrintList([]string{"name", "api_url", "token"}, []map[string]interface{}{}, nil)
				}
				return nil
			}

			columns := []string{"name", "api_url", "token"}
			rows := make([]map[string]interface{}, 0, len(globalConfig.Sites))

			for name, site := range globalConfig.Sites {
				tokenDisplay := maskToken(site.Token)
				rows = append(rows, map[string]interface{}{
					"name":    name,
					"api_url": site.APIURL,
					"token":   tokenDisplay,
				})
			}

			return printer.PrintList(columns, rows, nil)
		},
	}
}

// config init [--name=X] [--token=X] [--region=X | --api-url=X]
func newConfigInitCmd() *cobra.Command {
	var name, token, region, apiURL string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create stqry.yaml in the current directory",
		Example: `  # Pin current directory to a configured site
  stqry config init --name mysite

  # Initialise with inline credentials (no global config entry needed)
  stqry config init --token abc123 --region eu`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			// Full inline config: token + region/api-url provided — store credentials locally, skip global config.
			if token != "" {
				resolvedURL, err := resolveAPIURL(region, apiURL)
				if err != nil {
					return err
				}
				dirCfg := &config.DirectoryConfig{Token: token, APIURL: resolvedURL}
				if err := config.SaveDirectoryConfig(cwd, dirCfg); err != nil {
					return fmt.Errorf("saving directory config: %w", err)
				}
				if !flagQuiet && !flagJSON {
					fmt.Printf("Initialised stqry.yaml with inline site credentials.\n")
				}
				return nil
			}

			// Name-only: reference a site from global config.
			if name == "" {
				return fmt.Errorf("--name is required (or provide --token and --region/--api-url for inline config)")
			}

			if _, exists := globalConfig.Sites[name]; !exists {
				names := make([]string, 0, len(globalConfig.Sites))
				for n := range globalConfig.Sites {
					names = append(names, n)
				}
				hint := ""
				if len(names) > 0 {
					hint = fmt.Sprintf(" Available sites: %s", strings.Join(names, ", "))
				}
				return fmt.Errorf("site %q not found in global config.%s", name, hint)
			}

			dirCfg := &config.DirectoryConfig{Site: name}
			if err := config.SaveDirectoryConfig(cwd, dirCfg); err != nil {
				return fmt.Errorf("saving directory config: %w", err)
			}

			if !flagQuiet && !flagJSON {
				fmt.Printf("Initialised .stqry/config.yaml with site %q.\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Site name from global config to use in this directory")
	cmd.Flags().StringVar(&token, "token", "", "API token (stores credentials inline, skips global config)")
	cmd.Flags().StringVar(&region, "region", "", "Region (us, ca, eu, sg, au)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API base URL (overrides --region, e.g. for staging)")

	return cmd
}
