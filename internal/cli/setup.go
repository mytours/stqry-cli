package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mytours/stqry-cli/internal/buildinfo"
	"github.com/mytours/stqry-cli/internal/skills"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		Example: `  # Install Claude Code skill files into this project
  stqry setup claude`,
	}
	cmd.AddCommand(newSetupClaudeCmd())
	return cmd
}

func newSetupClaudeCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Install Claude Code skill files",
		Long: "Install STQRY CLI skill files for AI-assisted workflows. " +
			"Re-running this command always overwrites existing files — use it to update stale skills.",
		Example: `  # Install into .claude/commands/ for this project (Claude Code)
  stqry setup claude

  # Install globally into ~/.claude/commands/ (Claude Code)
  stqry setup claude --global`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetDir string

			switch {
			case global:
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("resolving home directory: %w", err)
				}
				targetDir = filepath.Join(home, ".claude", "commands")
			default:
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				targetDir = filepath.Join(cwd, ".claude", "commands")
			}

			names, err := skills.EmbeddedSkillNames()
			if err != nil {
				return fmt.Errorf("reading embedded skills: %w", err)
			}

			if err := skills.InstallAll(targetDir, skills.LayoutCode, buildinfo.Version); err != nil {
				return err
			}

			for _, name := range names {
				fmt.Printf("Installed %s\n", filepath.Join(targetDir, name))
			}

			fmt.Printf("\n%d skill file(s) installed to %s\n", len(names), targetDir)
			fmt.Println("Restart Claude Code (or reload commands) to activate the new skills.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Install to ~/.claude/commands/ instead of ./.claude/commands/")
	return cmd
}
