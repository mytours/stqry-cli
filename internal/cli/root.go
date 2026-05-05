package cli

import (
	"fmt"
	"os"

	"github.com/itchyny/gojq"
	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/buildinfo"
	"github.com/mytours/stqry-cli/internal/config"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagSite     string
	flagLang     string
	flagJSON     bool
	flagQuiet    bool
	flagJQ       string
	flagProgress bool

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
	case "__complete", "__completeNoDesc",
		"completion", "config", "doctor", "help", "mcp", "setup", "skill":
		return true
	}
	return false
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "stqry",
		Short: "STQRY CLI — manage your STQRY sites",
		Long:  "stqry is a command-line tool for managing STQRY content: sites, collections, screens, media, and more.",
		Example: `  # Show version
  stqry --version

  # List all collections on the default site
  stqry collections list

  # Use a specific site for any command
  stqry --site mysite collections list`,
		Version: buildinfo.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 1. Initialise printer.
			printer = &output.Printer{JSON: flagJSON, Quiet: flagQuiet}

			if flagJQ != "" {
				query, err := gojq.Parse(flagJQ)
				if err != nil {
					return fmt.Errorf("invalid jq expression: %w", err)
				}
				code, err := gojq.Compile(query)
				if err != nil {
					return fmt.Errorf("invalid jq expression: %w", err)
				}
				printer.JQCode = code
			}

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

	rootCmd.SetVersionTemplate("stqry {{.Version}}\n")

	// Global flags.
	rootCmd.PersistentFlags().StringVar(&flagSite, "site", "", "Site name to use (overrides directory config)")
	rootCmd.PersistentFlags().StringVar(&flagLang, "lang", "", "Language code for content (e.g. en, fr)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Output minimal JSON (no envelope)")
	rootCmd.PersistentFlags().StringVar(&flagJQ, "jq", "", "Filter output with a jq expression (overrides --quiet)")
	rootCmd.PersistentFlags().BoolVar(&flagProgress, "progress", false, "Show upload progress on stderr (off by default; modelled on dd(1) status=progress)")

	// Subcommands.
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newCollectionsCmd())
	rootCmd.AddCommand(newScreensCmd())
	rootCmd.AddCommand(newMediaCmd())
	rootCmd.AddCommand(newUploadedFilesCmd())
	rootCmd.AddCommand(newProjectsCmd())
	rootCmd.AddCommand(newCodesCmd())
	rootCmd.AddCommand(newSetupCmd())
	rootCmd.AddCommand(newSkillCmd())
	rootCmd.AddCommand(newMCPCmd())
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newCompletionCmd())

	return rootCmd
}

// Execute builds the root command and runs it.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
