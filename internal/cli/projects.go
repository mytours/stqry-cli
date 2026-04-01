package cli

import "github.com/spf13/cobra"

func newProjectsCmd() *cobra.Command {
	return &cobra.Command{Use: "projects", Short: "Manage projects"}
}
