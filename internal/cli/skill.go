package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mytours/stqry-cli/internal/buildinfo"
	"github.com/mytours/stqry-cli/internal/skills"
	"github.com/spf13/cobra"
)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Inspect and export embedded skill files",
	}
	cmd.AddCommand(newSkillExportCmd())
	return cmd
}

func newSkillExportCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a Claude Desktop skill zip to the current directory",
		Long: "Build and write a stqry-skill.zip containing SKILL.md, SETUP.md, REFERENCE.md, " +
			"and WORKFLOWS.md. Install the zip via Claude Desktop Settings → Customise → Skills.",
		Example: `  # Export to current directory (./stqry-skill.zip)
  stqry skill export

  # Export to a specific directory
  stqry skill export --dir ~/Downloads`,
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir := dir
			if outDir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				outDir = cwd
			}

			zipBytes, err := skills.BuildZip(buildinfo.Version)
			if err != nil {
				return fmt.Errorf("building skill zip: %w", err)
			}

			outPath := filepath.Join(outDir, "stqry-skill.zip")
			if err := os.WriteFile(outPath, zipBytes, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", outPath, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Exported skill to %s\n\n", outPath)
			fmt.Fprintln(cmd.OutOrStdout(), "To install in Claude Desktop:")
			fmt.Fprintln(cmd.OutOrStdout(), "  1. Open Claude Desktop")
			fmt.Fprintln(cmd.OutOrStdout(), "  2. Go to Settings → Customise → Skills")
			fmt.Fprintln(cmd.OutOrStdout(), `  3. Click "Add Skill" and select stqry-skill.zip`)
			fmt.Fprintln(cmd.OutOrStdout(), "  4. Restart Claude Desktop to activate")
			fmt.Fprintf(cmd.OutOrStdout(), "\nSkill version: %s\n", buildinfo.Version)
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "Directory to write stqry-skill.zip (default: current directory)")
	return cmd
}
