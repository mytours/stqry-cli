package cli

import (
	"strconv"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
)

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects",
		Example: `  # List all projects
  stqry projects list

  # Get a project by ID
  stqry projects get 1`,
	}

	cmd.AddCommand(newProjectsListCmd())
	cmd.AddCommand(newProjectsGetCmd())

	return cmd
}

func newProjectsListCmd() *cobra.Command {
	var page, perPage int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Example: `  # List all projects
  stqry projects list

  # List using a specific site
  stqry projects list --site mysite

  # Filter with built-in jq (no external jq needed)
  stqry projects list --jq '.[].name'

  # Pipe to external jq (alternative)
  stqry projects list --quiet | jq '.[].id'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := map[string]string{}
			if page > 0 {
				query["page"] = strconv.Itoa(page)
			}
			if perPage > 0 {
				query["per_page"] = strconv.Itoa(perPage)
			}

			projects, meta, err := api.ListProjects(activeClient, query)
			if err != nil {
				return err
			}

			var outMeta *output.Meta
			if meta != nil {
				outMeta = &output.Meta{
					Page:    meta.Page,
					PerPage: meta.PerPage,
					Total:   meta.Count,
				}
			}

			columns := []string{"id", "name"}
			return printer.PrintList(columns, projects, outMeta)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Results per page")

	return cmd
}

func newProjectsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a project by ID",
		Example: `  # Get a project by ID
  stqry projects get 1

  # Get project details as JSON
  stqry projects get 1 --json

  # Filter a specific field
  stqry projects get 1 --jq '.name'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := api.GetProject(activeClient, args[0])
			if err != nil {
				return err
			}
			return printer.PrintOne(project, &output.Meta{})
		},
	}
}

