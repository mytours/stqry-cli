package cli

import (
	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long:  "Generate shell completion scripts for bash, zsh, fish, or PowerShell.",
	}

	cmd.AddCommand(newCompletionBashCmd())
	cmd.AddCommand(newCompletionZshCmd())
	cmd.AddCommand(newCompletionFishCmd())
	cmd.AddCommand(newCompletionPowerShellCmd())

	return cmd
}

func newCompletionBashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		Long: `Generate bash completion script.

To load completions in your current shell session:
  source <(stqry completion bash)

To load completions for every new session (Linux):
  stqry completion bash > /etc/bash_completion.d/stqry

To load completions for every new session (macOS with bash-completion@2):
  stqry completion bash > $(brew --prefix)/etc/bash_completion.d/stqry`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		},
	}
}

func newCompletionZshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		Long: `Generate zsh completion script.

To load completions in your current shell session:
  source <(stqry completion zsh)

To load completions for every new session, add to ~/.zshrc:
  echo 'source <(stqry completion zsh)' >> ~/.zshrc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		},
	}
}

func newCompletionFishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		Long: `Generate fish completion script.

To load completions for every new session:
  stqry completion fish > ~/.config/fish/completions/stqry.fish`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}
}

func newCompletionPowerShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "powershell",
		Short: "Generate PowerShell completion script",
		Long: `Generate PowerShell completion script.

To load completions in your current session:
  stqry completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the above line to your PowerShell profile.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		},
	}
}
