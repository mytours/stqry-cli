package cli

import (
	"strconv"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
)

func newUploadedFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "uploaded-files",
		Aliases: []string{"files"},
		Short:   "Inspect uploaded file metadata (filename, size, hash, dimensions)",
		Long: `Uploaded files are the binaries that media items reference. Each media
item points at one or more uploaded_file records via file_uploaded_file_id
and thumbnail_uploaded_file_id (per language). This command surfaces the
underlying file metadata directly — useful for auditing what's in an
account, finding orphaned files, or building a manifest before bulk
operations.`,
		Example: `  # List uploaded files (paginate with --page / --per-page)
  stqry uploaded-files list --per-page 500

  # Show one uploaded file's metadata
  stqry uploaded-files get 359201

  # Total bytes across one page
  stqry uploaded-files list --per-page 500 --jq '[.[].file_size] | add'`,
	}
	cmd.AddCommand(newUploadedFilesListCmd())
	cmd.AddCommand(newUploadedFilesGetCmd())
	return cmd
}

func newUploadedFilesListCmd() *cobra.Command {
	var page, perPage int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uploaded files",
		Example: `  # List uploaded files (default API page size)
  stqry uploaded-files list

  # Walk every file (max per_page is 500; meta.pages tells you how many to fetch)
  stqry uploaded-files list --per-page 500 --jq '.[].filename'

  # Sum of every file's bytes on one page
  stqry uploaded-files list --per-page 500 --jq '[.[].file_size] | add'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := map[string]string{}
			if page > 0 {
				query["page"] = strconv.Itoa(page)
			}
			if perPage > 0 {
				query["per_page"] = strconv.Itoa(perPage)
			}

			files, meta, err := api.ListUploadedFiles(activeClient, query)
			if err != nil {
				printer.PrintError(err)
				return err
			}

			var outMeta *output.Meta
			if meta != nil {
				outMeta = &output.Meta{Page: meta.Page, PerPage: meta.PerPage, Total: meta.Count}
			}

			cols := []string{"id", "filename", "content_type", "file_size", "status"}
			return printer.PrintList(cols, files, outMeta)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Results per page")

	return cmd
}

func newUploadedFilesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get an uploaded file by ID",
		Example: `  # Get an uploaded file by ID
  stqry uploaded-files get 359201

  # Filter a specific field
  stqry uploaded-files get 359201 --jq '.filename'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := api.GetUploadedFile(activeClient, args[0])
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(f, nil)
		},
	}
	return cmd
}
