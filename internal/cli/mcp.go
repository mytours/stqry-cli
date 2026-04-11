package cli

import (
	"github.com/mytours/stqry-cli/internal/mcp"
	"github.com/spf13/cobra"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
	}
	cmd.AddCommand(newMCPServeCmd())
	return cmd
}

func newMCPServeCmd() *cobra.Command {
	var flagSite string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the STQRY MCP server (stdio transport)",
		Long:  "Starts an MCP server on stdio. Configure your MCP client to run: stqry mcp serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mcp.Serve(flagSite)
		},
	}
	cmd.Flags().StringVar(&flagSite, "site", "", "Site name to use (overrides STQRY_SITE and stqry.yaml)")
	return cmd
}
