package cli

import "github.com/spf13/cobra"

func newMediaCmd() *cobra.Command {
	return &cobra.Command{Use: "media", Short: "Manage media assets"}
}
