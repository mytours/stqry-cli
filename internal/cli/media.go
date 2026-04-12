package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
)

// validMediaTypes mirrors MediaItem::MEDIA_ITEM_SUBTYPES_SHORT in mytours-web
// (app/models/media_item.rb). Keep in sync if new subtypes are added.
var validMediaTypes = []string{"map", "webpackage", "animation", "audio", "image", "video", "webvideo", "ar", "data"}

func validateMediaType(t string) error {
	for _, v := range validMediaTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid media type %q (valid: %s)", t, strings.Join(validMediaTypes, ", "))
}

func newMediaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Manage media assets",
	}

	cmd.AddCommand(newMediaListCmd())
	cmd.AddCommand(newMediaGetCmd())
	cmd.AddCommand(newMediaCreateCmd())
	cmd.AddCommand(newMediaUpdateCmd())
	cmd.AddCommand(newMediaDeleteCmd())
	cmd.AddCommand(newMediaUploadCmd())

	return cmd
}

// media list [--page --per-page --q]
func newMediaListCmd() *cobra.Command {
	var page, perPage int
	var q string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List media items",
		RunE: func(cmd *cobra.Command, args []string) error {
			query := map[string]string{}
			if page > 0 {
				query["page"] = strconv.Itoa(page)
			}
			if perPage > 0 {
				query["per_page"] = strconv.Itoa(perPage)
			}
			if q != "" {
				query["q"] = q
			}
			if flagLang != "" {
				query["language"] = flagLang
			}

			items, meta, err := api.ListMediaItems(activeClient, query)
			if err != nil {
				printer.PrintError(err)
				return err
			}

			var outMeta *output.Meta
			if meta != nil {
				outMeta = &output.Meta{Page: meta.Page, PerPage: meta.PerPage, Total: meta.Count}
			}

			cols := []string{"id", "name", "type", "primary_language"}
			return printer.PrintList(cols, items, outMeta)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Items per page")
	cmd.Flags().StringVar(&q, "q", "", "Search query")

	return cmd
}

// media get <id>
func newMediaGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a media item by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			item, err := api.GetMediaItem(activeClient, args[0])
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}
}

// media create --file=<path> --type=X [--name=X] [--lang=X]
func newMediaCreateCmd() *cobra.Command {
	var filePath, mediaType, name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a media item (optionally uploading a file)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateMediaType(mediaType); err != nil {
				return err
			}

			fields := map[string]interface{}{
				"type": mediaType,
			}
			if name != "" {
				fields["name"] = name
			}

			// If a file is provided, upload it first.
			if filePath != "" {
				fmt.Printf("Uploading %s...\n", filePath)
				uploadedFile, err := api.UploadFile(activeClient, filePath, "", func(written, total int64) {
					if total > 0 {
						pct := float64(written) / float64(total) * 100
						fmt.Printf("\rUploading: %.0f%%", pct)
					}
				}, func(msg string) {
					fmt.Printf("\nProcessing: %s", msg)
				})
				fmt.Println()
				if err != nil {
					printer.PrintError(err)
					return err
				}

				uploadedFileID := ""
				if id, ok := uploadedFile["id"].(string); ok {
					uploadedFileID = id
				} else if id, ok := uploadedFile["id"].(float64); ok {
					uploadedFileID = strconv.Itoa(int(id))
				}

				lang := flagLang
				if lang != "" {
					fields["file_uploaded_file_id"] = map[string]interface{}{lang: uploadedFileID}
				} else {
					fields["file_uploaded_file_id"] = uploadedFileID
				}
			}

			item, err := api.CreateMediaItem(activeClient, fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to file to upload")
	cmd.Flags().StringVar(&mediaType, "type", "", fmt.Sprintf("Media item type (required; one of: %s)", strings.Join(validMediaTypes, ", ")))
	cmd.Flags().StringVar(&name, "name", "", "Media item name")
	cmd.MarkFlagRequired("type")

	return cmd
}

// media update <id> [--name=X]
func newMediaUpdateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a media item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := map[string]interface{}{}
			if name != "" {
				fields["name"] = name
			}

			item, err := api.UpdateMediaItem(activeClient, args[0], fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New name for the media item")

	return cmd
}

// media delete <id>
func newMediaDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a media item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var query map[string]string
			if flagLang != "" {
				query = map[string]string{"language": flagLang}
			}

			if err := api.DeleteMediaItem(activeClient, args[0], query); err != nil {
				printer.PrintError(err)
				return err
			}
			if !flagJSON && !flagQuiet {
				fmt.Printf("Media item %s deleted.\n", args[0])
			}
			return nil
		},
	}
}

// media upload <file> [--lang=X] [--media-id=X]
func newMediaUploadCmd() *cobra.Command {
	var mediaID string

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a file (optionally attach to a media item)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			uploadedFile, err := api.UploadFile(activeClient, filePath, "", func(written, total int64) {
				if total > 0 {
					pct := float64(written) / float64(total) * 100
					fmt.Printf("\rUploading: %.0f%%", pct)
				}
			}, func(msg string) {
				fmt.Printf("\nProcessing: %s", msg)
			})
			fmt.Println()
			if err != nil {
				printer.PrintError(err)
				return err
			}

			// If --media-id set, PATCH the media item with the uploaded file.
			if mediaID != "" {
				lang := flagLang
				if lang == "" {
					return fmt.Errorf("--lang (or --lang global flag) is required when --media-id is set")
				}

				uploadedFileID := ""
				if id, ok := uploadedFile["id"].(string); ok {
					uploadedFileID = id
				} else if id, ok := uploadedFile["id"].(float64); ok {
					uploadedFileID = strconv.Itoa(int(id))
				}

				fields := map[string]interface{}{
					"file_uploaded_file_id": map[string]interface{}{lang: uploadedFileID},
				}
				item, err := api.UpdateMediaItem(activeClient, mediaID, fields)
				if err != nil {
					printer.PrintError(err)
					return err
				}
				return printer.PrintOne(item, nil)
			}

			return printer.PrintOne(uploadedFile, nil)
		},
	}

	cmd.Flags().StringVar(&mediaID, "media-id", "", "Attach uploaded file to this media item ID")

	return cmd
}
