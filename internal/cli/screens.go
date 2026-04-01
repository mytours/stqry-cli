package cli

import "github.com/spf13/cobra"

func newScreensCmd() *cobra.Command {
	return &cobra.Command{Use: "screens", Short: "Manage screens"}
}
