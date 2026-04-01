package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mytours/stqry-cli/internal/skills"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
	}
	cmd.AddCommand(newSetupClaudeCmd())
	return cmd
}

func newSetupClaudeCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Install Claude Code skill files",
		Long:  "Install STQRY CLI skill files into the Claude Code commands directory for AI-assisted workflows.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine target directory.
			var targetDir string
			if global {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("resolving home directory: %w", err)
				}
				targetDir = filepath.Join(home, ".claude", "commands")
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				targetDir = filepath.Join(cwd, ".claude", "commands")
			}

			// Create target directory.
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", targetDir, err)
			}

			// Read and write each embedded .md file.
			entries, err := skills.SkillFiles.ReadDir(".")
			if err != nil {
				return fmt.Errorf("reading embedded skills: %w", err)
			}

			installed := 0
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				data, err := skills.SkillFiles.ReadFile(entry.Name())
				if err != nil {
					return fmt.Errorf("reading embedded file %s: %w", entry.Name(), err)
				}

				dest := filepath.Join(targetDir, entry.Name())
				if err := os.WriteFile(dest, data, 0o644); err != nil {
					return fmt.Errorf("writing %s: %w", dest, err)
				}

				fmt.Printf("Installed %s\n", dest)
				installed++
			}

			fmt.Printf("\n%d skill file(s) installed to %s\n", installed, targetDir)
			fmt.Println("Restart Claude Code (or reload commands) to activate the new skills.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Install to ~/.claude/commands/ instead of ./.claude/commands/")
	return cmd
}
