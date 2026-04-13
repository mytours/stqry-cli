package cli

import (
	"fmt"
	"strings"

	"github.com/mytours/stqry-cli/internal/buildinfo"
	"github.com/mytours/stqry-cli/internal/skills"
	"github.com/spf13/cobra"
)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Inspect embedded skill files",
	}
	cmd.AddCommand(newSkillDumpCmd())
	return cmd
}

func newSkillDumpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dump [skill-name]",
		Short: "Print embedded skill content to stdout",
		Long: "Print a skill's content (with frontmatter) to stdout. " +
			"With no argument, lists available skill names. " +
			"Useful for piping to a file for manual installation.",
		Example: `  # List available skills
  stqry skill dump

  # Print a skill to stdout
  stqry skill dump stqry-reference

  # Pipe to a file for manual installation
  stqry skill dump stqry-reference > stqry-reference.md`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := skills.EmbeddedSkillNames()
			if err != nil {
				return fmt.Errorf("reading embedded skills: %w", err)
			}

			if len(args) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Available skills:")
				for _, name := range names {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", strings.TrimSuffix(name, ".md"))
				}
				return nil
			}

			target := args[0]
			// Accept with or without .md suffix.
			filename := target
			if !strings.HasSuffix(filename, ".md") {
				filename += ".md"
			}

			data, err := skills.SkillFiles.ReadFile(filename)
			if err != nil {
				available := make([]string, len(names))
				for i, n := range names {
					available[i] = strings.TrimSuffix(n, ".md")
				}
				return fmt.Errorf("skill %q not found (available: %s)", target, strings.Join(available, ", "))
			}

			content := skills.BuildFrontmatter(buildinfo.Version, data)
			cmd.OutOrStdout().Write(content) //nolint:errcheck
			return nil
		},
	}
}
