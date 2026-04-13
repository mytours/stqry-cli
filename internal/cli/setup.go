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
	var desktop bool
	var dir string

	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Install Claude Code skill files",
		Long: "Install STQRY CLI skill files for AI-assisted workflows. " +
			"Re-running always overwrites existing files — use it to update stale skills.\n\n" +
			"For Claude Desktop, use `stqry skill export` to produce a stqry-skill.zip, " +
			"then install it via Claude Desktop Settings → Customise → Skills.",
		Example: `  # Install into .claude/commands/ for this project (Claude Code)
  stqry setup claude

  # Install globally into ~/.claude/commands/ (Claude Code)
  stqry setup claude --global

  # Export a skill zip for Claude Desktop
  stqry skill export`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetDir string

			switch {
			case desktop:
				fmt.Fprintln(cmd.ErrOrStderr(), "Warning: --desktop exports flat .md files, not a zip. Use `stqry skill export` instead for Claude Desktop.")
				exportDir := dir
				if exportDir == "" {
					home, err := os.UserHomeDir()
					if err != nil {
						return fmt.Errorf("resolving home directory: %w", err)
					}
					exportDir = filepath.Join(home, "Downloads")
				}
				targetDir = exportDir
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

			if err := skills.InstallAll(targetDir, buildinfo.Version); err != nil {
				return err
			}

			for _, name := range names {
				if desktop {
					fmt.Fprintf(cmd.OutOrStdout(), "Exported %s\n", filepath.Join(targetDir, name))
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Installed %s\n", filepath.Join(targetDir, name))
				}
			}

			if desktop {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%d skill file(s) exported to %s\n", len(names), targetDir)
				fmt.Fprintln(cmd.OutOrStdout(), "\nTo install in Claude Desktop:")
				fmt.Fprintln(cmd.OutOrStdout(), "  1. Open Claude Desktop")
				fmt.Fprintln(cmd.OutOrStdout(), "  2. Go to Settings → Skills")
				fmt.Fprintln(cmd.OutOrStdout(), `  3. Click "Add Skill" for each file above`)
				fmt.Fprintln(cmd.OutOrStdout(), "  4. Restart Claude Desktop to activate")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%d skill file(s) installed to %s\n", len(names), targetDir)
				fmt.Fprintln(cmd.OutOrStdout(), "Restart Claude Code (or reload commands) to activate the new skills.")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Install to ~/.claude/commands/ instead of ./.claude/commands/")
	cmd.Flags().BoolVar(&desktop, "desktop", false, "Export skill files for manual install via Claude Desktop Settings → Skills")
	cmd.Flags().StringVar(&dir, "dir", "", "Export directory for --desktop (default: ~/Downloads)")
	cmd.MarkFlagsMutuallyExclusive("global", "desktop")
	return cmd
}
