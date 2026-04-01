package cli

import "github.com/spf13/cobra"

func newCodesCmd() *cobra.Command {
	return &cobra.Command{Use: "codes", Short: "Manage QR/NFC codes"}
}
