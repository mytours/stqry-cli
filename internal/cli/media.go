package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func validateMediaType(t string) error {
	for _, v := range api.ValidMediaTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid media type %q (valid: %s)", t, strings.Join(api.ValidMediaTypes, ", "))
}

func newMediaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Manage media assets",
		Example: `  # List all media items
  stqry media list

  # Upload an image and create a media item
  stqry media create --type image --file ./photo.jpg`,
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
		Example: `  # List all media items
  stqry media list

  # Search for media by name
  stqry media list --q "banner"

  # List using a specific site
  stqry media list --site mysite

  # Filter with built-in jq (no external jq needed)
  stqry media list --jq '.[].name'

  # Pipe to external jq (alternative)
  stqry media list --quiet | jq '.[].id'`,
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
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a media item by ID",
		Example: `  # Get a media item by ID
  stqry media get 55

  # Get media item details as JSON
  stqry media get 55 --json

  # Filter a specific field
  stqry media get 55 --jq '.name'`,
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
	cmd.ValidArgsFunction = completeMediaIDs
	return cmd
}

// media create --file=<path> --type=X [--name=X] [--lang=X] [--caption=X] [--attribution=X] [--description=X] [--title=X] [--transcription=X] [--thumbnail-media-item-id=X] [--compress=BOOL]
func newMediaCreateCmd() *cobra.Command {
	var filePath, mediaType, name string
	var caption, attribution, description, title, transcription string
	var thumbnailMediaItemID int
	var compress bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a media item (optionally uploading a file)",
		Example: `  # Create an image media item with a file
  stqry media create --type image --file ./photo.jpg

  # Create an image with a caption and CC attribution on the media item itself
  stqry media create --type image --file ./photo.jpg \
    --caption "Court Street subway entrance, c. 1928" \
    --attribution "Rochester Subway Archive · public domain"

  # Create a video media item with a name and language
  stqry media create --type video --file ./tour.mp4 --name "City Tour" --lang en

  # Create an audio item with a transcription (accessibility)
  stqry media create --type audio --file ./stop_1.mp3 --transcription "Welcome to the tour..."

  # Create an audio item with an image thumbnail (poster) — the referenced
  # image media item shows up wherever the audio is rendered
  stqry media create --type audio --file ./stop_1.mp3 --title "Stop 1" --thumbnail-media-item-id 99

  # Create an image and prefer the uncropped original (skip transform/compression)
  stqry media create --type image --file ./photo.jpg --compress=false`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateMediaType(mediaType); err != nil {
				return err
			}
			lang := flagLang
			if lang == "" {
				lang = "en"
			}

			fields := map[string]interface{}{
				"type": mediaType,
			}
			if name != "" {
				fields["name"] = name
			}
			if caption != "" {
				fields["caption"] = map[string]interface{}{lang: caption}
			}
			if attribution != "" {
				fields["attribution"] = map[string]interface{}{lang: attribution}
			}
			if description != "" {
				fields["description"] = map[string]interface{}{lang: description}
			}
			if title != "" {
				fields["title"] = map[string]interface{}{lang: title}
			}
			if transcription != "" {
				fields["transcription"] = map[string]interface{}{lang: transcription}
			}
			if cmd.Flags().Changed("thumbnail-media-item-id") {
				fields["thumbnail_media_item_id"] = thumbnailMediaItemID
			}
			if cmd.Flags().Changed("compress") {
				fields["compress_media"] = compress
			}

			// If a file is provided, upload it first.
			if filePath != "" {
				// Progress goes to stderr so --json/--quiet/--jq output on stdout stays
				// parseable. Opt-in via --progress (off by default, dd(1)-style).
				showProgress := progressEnabled()
				var onProgress func(written, total int64)
				var onStatus func(string)
				if showProgress {
					fmt.Fprintf(os.Stderr, "Uploading %s...\n", filePath)
					onProgress = func(written, total int64) {
						if total > 0 {
							pct := float64(written) / float64(total) * 100
							fmt.Fprintf(os.Stderr, "\rUploading: %.0f%%", pct)
						}
					}
					onStatus = func(msg string) {
						fmt.Fprintf(os.Stderr, "\nProcessing: %s", msg)
					}
				}
				uploadedFile, err := api.UploadFile(activeClient, filePath, "", onProgress, onStatus)
				if showProgress {
					fmt.Fprintln(os.Stderr)
				}
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

				if flagLang != "" {
					fields["file_uploaded_file_id"] = map[string]interface{}{flagLang: uploadedFileID}
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
	cmd.Flags().StringVar(&mediaType, "type", "", fmt.Sprintf("Media item type (required; one of: %s)", strings.Join(api.ValidMediaTypes, ", ")))
	cmd.Flags().StringVar(&name, "name", "", "Media item name (internal label; not shown to end users)")
	cmd.Flags().StringVar(&caption, "caption", "", "Image caption (image media items)")
	cmd.Flags().StringVar(&attribution, "attribution", "", "Credit / attribution line (image media items — e.g. \"Photo by X · CC BY 4.0\")")
	cmd.Flags().StringVar(&description, "description", "", "Long description (image media items)")
	cmd.Flags().StringVar(&title, "title", "", "Display title (audio media items)")
	cmd.Flags().StringVar(&transcription, "transcription", "", "Transcription (audio media items — accessibility)")
	cmd.Flags().IntVar(&thumbnailMediaItemID, "thumbnail-media-item-id", 0, "Image media item ID to use as the thumbnail / poster (audio / video media items)")
	cmd.Flags().BoolVar(&compress, "compress", true, "Whether to compress / transform the uploaded media. Pass --compress=false to prefer the uncropped original (\"don't transform\" / \"prefer uncropped image\")")
	cmd.MarkFlagRequired("type")

	return cmd
}

// media update <id> [--name=X] [--caption=X] [--attribution=X] [--description=X] [--title=X] [--transcription=X] [--thumbnail-media-item-id=X] [--compress=BOOL]
func newMediaUpdateCmd() *cobra.Command {
	var name, caption, attribution, description, title, transcription string
	var thumbnailMediaItemID int
	var compress bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a media item",
		Example: `  # Rename a media item
  stqry media update 55 --name "New Banner"

  # Set caption + attribution on an image (credits belong on the MediaItem, not the section)
  stqry media update 55 --caption "Subway tunnel, 2001" --attribution "Photo by A. Borchert · CC BY-SA 4.0"

  # Set transcription on an audio item for accessibility
  stqry media update 72 --transcription "Welcome to the tour..." --lang en

  # Attach an image as the thumbnail / poster for an audio item
  stqry media update 72 --thumbnail-media-item-id 99

  # Prefer the uncropped original — skip transform/compression
  stqry media update 55 --compress=false

  # Clear a field (pass empty string)
  stqry media update 55 --attribution ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			lang := flagLang
			if lang == "" {
				lang = "en"
			}
			fields := map[string]interface{}{}
			cmd.Flags().Visit(func(f *flag.Flag) {
				switch f.Name {
				case "name":
					fields["name"] = name
				case "caption":
					fields["caption"] = map[string]interface{}{lang: caption}
				case "attribution":
					fields["attribution"] = map[string]interface{}{lang: attribution}
				case "description":
					fields["description"] = map[string]interface{}{lang: description}
				case "title":
					fields["title"] = map[string]interface{}{lang: title}
				case "transcription":
					fields["transcription"] = map[string]interface{}{lang: transcription}
				case "thumbnail-media-item-id":
					fields["thumbnail_media_item_id"] = thumbnailMediaItemID
				case "compress":
					fields["compress_media"] = compress
				}
			})

			item, err := api.UpdateMediaItem(activeClient, args[0], fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Media item name (internal label; not shown to end users)")
	cmd.Flags().StringVar(&caption, "caption", "", "Image caption (image media items)")
	cmd.Flags().StringVar(&attribution, "attribution", "", "Credit / attribution line (image media items — e.g. \"Photo by X · CC BY 4.0\")")
	cmd.Flags().StringVar(&description, "description", "", "Long description (image media items)")
	cmd.Flags().StringVar(&title, "title", "", "Display title (audio media items)")
	cmd.Flags().StringVar(&transcription, "transcription", "", "Transcription (audio media items — accessibility)")
	cmd.Flags().IntVar(&thumbnailMediaItemID, "thumbnail-media-item-id", 0, "Image media item ID to use as the thumbnail / poster (audio / video media items; pass 0 to clear)")
	cmd.Flags().BoolVar(&compress, "compress", true, "Whether to compress / transform the media. Pass --compress=false to prefer the uncropped original (\"don't transform\" / \"prefer uncropped image\")")
	cmd.ValidArgsFunction = completeMediaIDs

	return cmd
}

// media delete <id>
func newMediaDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a media item",
		Example: `  # Delete a media item
  stqry media delete 55

  # Delete a language variant of a media item
  stqry media delete 55 --lang fr`,
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
	cmd.ValidArgsFunction = completeMediaIDs
	return cmd
}

// media upload <file> [--lang=X] [--media-id=X]
func newMediaUploadCmd() *cobra.Command {
	var mediaID string

	cmd := &cobra.Command{
		Use:   "upload <file> --media-id <id>",
		Short: "Attach a new file to an existing media item (use 'media create' to create a new one)",
		Long: `Upload a file and attach it to an existing media item via --media-id.

Almost always use ` + "`stqry media create`" + ` instead. ` + "`stqry media upload`" + ` is
only for attaching a new file (e.g. a language variant or replacement) to a
media item that already exists.

--media-id is required. Running without it is not supported because the
resulting uploaded file would be orphaned — invisible in STQRY Builder and
unlinkable from the CLI afterwards.`,
		Example: `  # PREFERRED: create a new media item with a file
  stqry media create --type image --file ./photo.jpg

  # Attach a new file to an EXISTING media item (e.g. add a language variant)
  stqry media upload ./photo.jpg --media-id 55 --lang en`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if mediaID == "" {
				return fmt.Errorf("--media-id is required. To upload a file and create a new media item for it, use `stqry media create --type <type> --file <path>` instead.\n\nTo attach a file to an existing media item, pass --media-id <id> --lang <code>")
			}

			lang := flagLang
			if lang == "" {
				return fmt.Errorf("--lang (or --lang global flag) is required")
			}

			filePath := args[0]

			// Progress goes to stderr so --json/--quiet/--jq output on stdout stays
			// parseable. Opt-in via --progress (off by default, dd(1)-style).
			showProgress := progressEnabled()
			var onProgress func(written, total int64)
			var onStatus func(string)
			if showProgress {
				fmt.Fprintf(os.Stderr, "Uploading %s...\n", filePath)
				onProgress = func(written, total int64) {
					if total > 0 {
						pct := float64(written) / float64(total) * 100
						fmt.Fprintf(os.Stderr, "\rUploading: %.0f%%", pct)
					}
				}
				onStatus = func(msg string) {
					fmt.Fprintf(os.Stderr, "\nProcessing: %s", msg)
				}
			}
			uploadedFile, err := api.UploadFile(activeClient, filePath, "", onProgress, onStatus)
			if showProgress {
				fmt.Fprintln(os.Stderr)
			}
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

			fields := map[string]interface{}{
				"file_uploaded_file_id": map[string]interface{}{lang: uploadedFileID},
			}
			item, err := api.UpdateMediaItem(activeClient, mediaID, fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&mediaID, "media-id", "", "Attach uploaded file to this media item ID (required)")

	return cmd
}
