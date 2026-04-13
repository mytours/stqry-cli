package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/completion"
	"github.com/mytours/stqry-cli/internal/config"
	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long:  "Generate shell completion scripts for bash, zsh, fish, or PowerShell.",
	}

	cmd.AddCommand(newCompletionBashCmd())
	cmd.AddCommand(newCompletionZshCmd())
	cmd.AddCommand(newCompletionFishCmd())
	cmd.AddCommand(newCompletionPowerShellCmd())
	cmd.AddCommand(newCompletionRefreshCmd())
	cmd.AddCommand(newCompletionStatusCmd())

	return cmd
}

func newCompletionBashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		Long: `Generate bash completion script.

To load completions in your current shell session:
  source <(stqry completion bash)

To load completions for every new session (Linux):
  stqry completion bash > /etc/bash_completion.d/stqry

To load completions for every new session (macOS with bash-completion@2):
  stqry completion bash > $(brew --prefix)/etc/bash_completion.d/stqry`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		},
	}
}

func newCompletionZshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		Long: `Generate zsh completion script.

To load completions in your current shell session:
  source <(stqry completion zsh)

To load completions for every new session, add to ~/.zshrc:
  echo 'source <(stqry completion zsh)' >> ~/.zshrc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		},
	}
}

func newCompletionFishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		Long: `Generate fish completion script.

To load completions for every new session:
  stqry completion fish > ~/.config/fish/completions/stqry.fish`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}
}

func newCompletionPowerShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "powershell",
		Short: "Generate PowerShell completion script",
		Long: `Generate PowerShell completion script.

To load completions in your current session:
  stqry completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the above line to your PowerShell profile.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		},
	}
}

func newCompletionRefreshCmd() *cobra.Command {
	var siteName string

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh the local completion cache",
		Long:  "Fetch all resource names from the API and update the local completion cache.",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := siteName
			if name == "" {
				name = flagSite
			}

			cfg, err := config.LoadGlobalConfig(config.DefaultGlobalConfigPath())
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			cwd, _ := os.Getwd()
			dirCfg, _ := config.FindDirectoryConfig(cwd)

			if name == "" && dirCfg != nil {
				name = dirCfg.Site
			}

			if name == "" {
				return fmt.Errorf("no site specified; use --site or run from a directory with stqry.yaml")
			}

			site, err := config.ResolveSite(cfg, name, dirCfg)
			if err != nil {
				return err
			}

			client := api.NewClient(site.APIURL, site.Token)

			resources := []struct {
				key   string
				fetch func() ([]completion.CacheEntry, error)
			}{
				{"collections", func() ([]completion.CacheEntry, error) {
					return fetchAllEntries(func(page int) ([]map[string]interface{}, *api.PaginationMeta, error) {
						return api.ListCollections(client, map[string]string{"page": strconv.Itoa(page), "per_page": "100"})
					})
				}},
				{"screens", func() ([]completion.CacheEntry, error) {
					return fetchAllEntries(func(page int) ([]map[string]interface{}, *api.PaginationMeta, error) {
						return api.ListScreens(client, map[string]string{"page": strconv.Itoa(page), "per_page": "100"})
					})
				}},
				{"media", func() ([]completion.CacheEntry, error) {
					return fetchAllEntries(func(page int) ([]map[string]interface{}, *api.PaginationMeta, error) {
						return api.ListMediaItems(client, map[string]string{"page": strconv.Itoa(page), "per_page": "100"})
					})
				}},
				{"projects", func() ([]completion.CacheEntry, error) {
					return fetchAllEntries(func(page int) ([]map[string]interface{}, *api.PaginationMeta, error) {
						return api.ListProjects(client, map[string]string{"page": strconv.Itoa(page), "per_page": "100"})
					})
				}},
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Refreshed completions for site %q:\n", name)
			for _, r := range resources {
				entries, err := r.fetch()
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  %-15s error: %v\n", r.key, err)
					continue
				}
				if err := completion.Save(name, r.key, entries); err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "  %-15s save error: %v\n", r.key, err)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %-15s %d items\n", r.key, len(entries))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&siteName, "site", "", "Site to refresh (defaults to active site)")
	return cmd
}

func newCompletionStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show completion cache status",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := resolveSiteNameForCompletion()
			if name == "" {
				return fmt.Errorf("no site specified; use --site or run from a directory with stqry.yaml")
			}

			resources := []string{"collections", "screens", "media", "projects"}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "%-15s %-8s %-12s %s\n", "Resource", "Items", "Age", "Status")
			fmt.Fprintf(w, "%-15s %-8s %-12s %s\n", "--------", "-----", "---", "------")

			for _, resource := range resources {
				items, stale, err := completion.Load(name, resource)
				if err != nil || (len(items) == 0 && stale) {
					fmt.Fprintf(w, "%-15s %-8s %-12s %s\n", resource, "-", "-", "not cached")
					continue
				}
				path, _ := completion.CachePath(name, resource)
				info, err := os.Stat(path)
				age := "-"
				if err == nil {
					d := time.Since(info.ModTime())
					age = formatAge(d)
				}
				status := "fresh"
				if stale {
					status = "stale"
				}
				fmt.Fprintf(w, "%-15s %-8d %-12s %s\n", resource, len(items), age, status)
			}
			return nil
		},
	}
}

func formatAge(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}

// fetchAllEntries pages through a list endpoint collecting all IDs and names.
func fetchAllEntries(listFn func(page int) ([]map[string]interface{}, *api.PaginationMeta, error)) ([]completion.CacheEntry, error) {
	var entries []completion.CacheEntry
	for page := 1; ; page++ {
		items, meta, err := listFn(page)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			e := completion.CacheEntry{}
			switch v := item["id"].(type) {
			case float64:
				e.ID = strconv.Itoa(int(v))
			case string:
				e.ID = v
			}
			if name, ok := item["name"].(string); ok {
				e.Name = name
			}
			if e.ID != "" {
				entries = append(entries, e)
			}
		}
		if meta == nil || page >= meta.Pages {
			break
		}
	}
	return entries, nil
}
