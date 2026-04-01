package cli

import "github.com/spf13/cobra"

func newSetupCmd() *cobra.Command {
	return &cobra.Command{Use: "setup", Short: "Interactive setup wizard"}
}
