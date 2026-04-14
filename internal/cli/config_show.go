package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mytours/stqry-cli/internal/config"
	"github.com/spf13/cobra"
)

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the fully resolved configuration with source tracking",
		Example: `  # Show resolved config for the current directory
  stqry config show

  # Output as JSON
  stqry config show --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow()
		},
	}
}

func runConfigShow() error {
	configPath := config.DefaultGlobalConfigPath()

	globalCfg, err := config.LoadGlobalConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading global config: %w", err)
	}

	cwd, _ := os.Getwd()
	var dirCfg *config.DirectoryConfig
	var dirCfgPath string
	var dirCfgErr error
	if cwd != "" {
		dirCfg, dirCfgPath, dirCfgErr = config.FindDirectoryConfigWithPath(cwd)
	}

	// Best-effort resolution — no error if nothing configured.
	site, source, _ := config.ResolveSiteWithSource(globalCfg, flagSite, dirCfg)

	if flagJSON || flagQuiet {
		return printConfigShowJSON(configPath, dirCfgPath, dirCfg, dirCfgErr, globalCfg, site, source)
	}
	return printConfigShowHuman(configPath, dirCfgPath, dirCfg, dirCfgErr, globalCfg, site, source)
}

func printConfigShowHuman(configPath, dirCfgPath string, dirCfg *config.DirectoryConfig, dirCfgErr error, globalCfg *config.GlobalConfig, site *config.Site, source string) error {
	fmt.Println("Config files:")
	fmt.Printf("  Global:    %s (%d sites)\n", configPath, len(globalCfg.Sites))
	if dirCfgPath != "" {
		if dirCfgErr != nil {
			fmt.Printf("  Directory: %s (parse error: %s)\n", dirCfgPath, dirCfgErr)
		} else {
			fmt.Printf("  Directory: %s (%s)\n", dirCfgPath, dirConfigDescription(dirCfg))
		}
	} else {
		fmt.Printf("  Directory: (none)\n")
	}

	fmt.Println()
	if site != nil {
		region := regionFromURL(site.APIURL)
		fmt.Printf("Active site:\n")
		fmt.Printf("  Source:  %s\n", source)
		fmt.Printf("  API URL: %s\n", site.APIURL)
		fmt.Printf("  Token:   %s\n", maskToken(site.Token))
		if region != "" {
			fmt.Printf("  Region:  %s\n", region)
		}
	} else {
		fmt.Printf("Active site: (not configured)\n")
	}

	if len(globalCfg.Sites) > 0 {
		fmt.Println()
		fmt.Println("All sites:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  NAME\tAPI URL\tTOKEN")
		for name, s := range globalCfg.Sites {
			fmt.Fprintf(w, "  %s\t%s\t%s\n", name, s.APIURL, maskToken(s.Token))
		}
		w.Flush()
	}
	return nil
}

func printConfigShowJSON(configPath, dirCfgPath string, dirCfg *config.DirectoryConfig, dirCfgErr error, globalCfg *config.GlobalConfig, site *config.Site, source string) error {
	sites := make([]map[string]interface{}, 0, len(globalCfg.Sites))
	for name, s := range globalCfg.Sites {
		sites = append(sites, map[string]interface{}{
			"name":    name,
			"api_url": s.APIURL,
			"token":   maskToken(s.Token),
		})
	}

	var dirEntry interface{}
	if dirCfgPath != "" {
		if dirCfgErr != nil {
			dirEntry = map[string]interface{}{
				"path":  dirCfgPath,
				"error": dirCfgErr.Error(),
			}
		} else {
			dirEntry = map[string]interface{}{
				"path":    dirCfgPath,
				"content": dirConfigDescription(dirCfg),
			}
		}
	}

	var activeSite interface{}
	if site != nil {
		activeSite = map[string]interface{}{
			"source":  source,
			"api_url": site.APIURL,
			"token":   maskToken(site.Token),
			"region":  regionFromURL(site.APIURL),
		}
	}

	data := map[string]interface{}{
		"config_files": map[string]interface{}{
			"global": map[string]interface{}{
				"path":       configPath,
				"site_count": len(globalCfg.Sites),
			},
			"directory": dirEntry,
		},
		"active_site": activeSite,
		"sites":       sites,
	}

	return printer.PrintOne(data, nil)
}

// maskToken returns the first 8 chars of token followed by "...",
// or the full token if it is 8 characters or fewer.
func maskToken(token string) string {
	if len(token) > 8 {
		return token[:8] + "..."
	}
	return token
}

// regionFromURL extracts the region code from a STQRY API URL hostname.
// Returns empty string for non-standard URLs.
func regionFromURL(apiURL string) string {
	host := hostFromURL(apiURL)
	if strings.HasPrefix(host, "api-") && strings.HasSuffix(host, ".stqry.com") {
		parts := strings.SplitN(host, ".", 2)
		return strings.TrimPrefix(parts[0], "api-")
	}
	return ""
}

// dirConfigDescription returns a short human-readable description of a directory config.
func dirConfigDescription(dirCfg *config.DirectoryConfig) string {
	if dirCfg == nil {
		return ""
	}
	if dirCfg.Site != "" {
		return fmt.Sprintf("site: %s", dirCfg.Site)
	}
	if dirCfg.Token != "" {
		return "inline credentials"
	}
	return "empty"
}
