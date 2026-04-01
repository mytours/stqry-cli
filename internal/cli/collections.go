package cli

import "github.com/spf13/cobra"

func newCollectionsCmd() *cobra.Command {
	return &cobra.Command{Use: "collections", Short: "Manage collections"}
}
