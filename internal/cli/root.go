package cli

import (
	"fmt"
	"os"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/config"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagSite   string
	flagLang   string
	flagJSON   bool
	flagQuiet  bool

	globalConfig *config.GlobalConfig
	activeClient *api.Client
	printer      *output.Printer
)

// skipSiteResolution returns true for commands that don't need an API client.
func skipSiteResolution(cmd *cobra.Command) bool {
	root := cmd
	for root.HasParent() {
		root = root.Parent()
	}
	// Walk up to find the top-level subcommand name.
	sub := cmd
	for sub.HasParent() && sub.Parent() != root {
		sub = sub.Parent()
	}
	name := sub.Name()
	switch name {
	case "config", "help", "setup":
		return true
	}
	return false
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "stqry",
		Short: "STQRY CLI — manage your STQRY sites",
		Long:  "stqry is a command-line tool for managing STQRY content: sites, collections, screens, media, and more.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 1. Initialise printer.
			printer = &output.Printer{JSON: flagJSON, Quiet: flagQuiet}

			// 2. Load global config.
			var err error
			globalConfig, err = config.LoadGlobalConfig(config.DefaultGlobalConfigPath())
			if err != nil {
				return fmt.Errorf("loading global config: %w", err)
			}

			// 3. Skip site resolution for config/help/setup commands.
			if skipSiteResolution(cmd) {
				return nil
			}

			// 4. Resolve site and create API client.
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			dirCfg, err := config.FindDirectoryConfig(cwd)
			if err != nil {
				return fmt.Errorf("finding directory config: %w", err)
			}

			site, err := config.ResolveSite(globalConfig, flagSite, dirCfg)
			if err != nil {
				return err
			}

			activeClient = api.NewClient(site.APIURL, site.Token)
			return nil
		},
	}

	// Global flags.
	rootCmd.PersistentFlags().StringVar(&flagSite, "site", "", "Site name to use (overrides directory config)")
	rootCmd.PersistentFlags().StringVar(&flagLang, "lang", "", "Language code for content (e.g. en, fr)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Output minimal JSON (no envelope)")

	// Subcommands.
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newCollectionsCmd())
	rootCmd.AddCommand(newScreensCmd())
	rootCmd.AddCommand(newMediaCmd())
	rootCmd.AddCommand(newProjectsCmd())
	rootCmd.AddCommand(newCodesCmd())
	rootCmd.AddCommand(newSetupCmd())

	return rootCmd
}

// Execute builds the root command and runs it.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
