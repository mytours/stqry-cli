package cli

import (
	"fmt"
	"os"

	"github.com/mytours/stqry-cli/internal/completion"
	"github.com/mytours/stqry-cli/internal/config"
	"github.com/spf13/cobra"
)

const staleHint = "# cache stale — run: stqry completion refresh"

// resolveSiteNameForCompletion returns the active site name for cache lookups.
// Returns "" if a site name cannot be determined — completions must not error.
func resolveSiteNameForCompletion() string {
	if flagSite != "" {
		return flagSite
	}
	if envSite := os.Getenv("STQRY_SITE"); envSite != "" {
		return envSite
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dirCfg, err := config.FindDirectoryConfig(cwd)
	if err != nil || dirCfg == nil || dirCfg.Site == "" {
		return ""
	}
	return dirCfg.Site
}

func completionResults(site, resource string) ([]string, cobra.ShellCompDirective) {
	if site == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	items, stale, err := completion.Load(site, resource)
	if err != nil || len(items) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	results := make([]string, 0, len(items)+1)
	for _, e := range items {
		results = append(results, fmt.Sprintf("%s\t%s", e.ID, e.Name))
	}
	if stale {
		results = append(results, staleHint)
	}
	return results, cobra.ShellCompDirectiveNoFileComp
}

func completeCollectionIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResults(resolveSiteNameForCompletion(), "collections")
}

func completeScreenIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResults(resolveSiteNameForCompletion(), "screens")
}

func completeMediaIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResults(resolveSiteNameForCompletion(), "media")
}

func completeProjectIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completionResults(resolveSiteNameForCompletion(), "projects")
}
