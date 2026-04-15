package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// validCollectionTypes mirrors Collection::SUBTYPES_SHORT in mytours-web
// (app/models/collection.rb). Keep in sync if new subtypes are added.
var validCollectionTypes = []string{"list", "tour", "organization", "menu", "search"}

func validateCollectionType(t string) error {
	for _, v := range validCollectionTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid collection type %q (valid: %s)", t, strings.Join(validCollectionTypes, ", "))
}

func newCollectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collections",
		Short: "Manage collections",
		Example: `  # List all collections
  stqry collections list

  # Create a tour collection
  stqry collections create --name city-tour --type tour`,
	}

	cmd.AddCommand(newCollectionsListCmd())
	cmd.AddCommand(newCollectionsGetCmd())
	cmd.AddCommand(newCollectionsCreateCmd())
	cmd.AddCommand(newCollectionsUpdateCmd())
	cmd.AddCommand(newCollectionsDeleteCmd())
	cmd.AddCommand(newCollectionsItemsCmd())

	return cmd
}

func newCollectionsListCmd() *cobra.Command {
	var page, perPage int
	var q string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List collections",
		Example: `  # List all collections
  stqry collections list

  # Search for collections matching a keyword
  stqry collections list --q "museum"

  # List using a specific site, paginated
  stqry collections list --site mysite --page 2 --per-page 25

  # Filter with built-in jq (no external jq needed)
  stqry collections list --jq '.[].name'

  # Pipe to external jq (alternative)
  stqry collections list --quiet | jq '.[].id'`,
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

			cols, meta, err := api.ListCollections(activeClient, query)
			if err != nil {
				printer.PrintError(err)
				return err
			}

			var m *output.Meta
			if meta != nil {
				m = &output.Meta{Page: meta.Page, PerPage: meta.PerPage, Total: meta.Count}
			}
			return printer.PrintList([]string{"id", "name", "title"}, cols, m)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Results per page")
	cmd.Flags().StringVar(&q, "q", "", "Search query")

	return cmd
}

func newCollectionsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a collection",
		Example: `  # Get a collection by ID
  stqry collections get 42

  # Get collection details as JSON
  stqry collections get 42 --json

  # Filter a specific field
  stqry collections get 42 --jq '.name'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, err := api.GetCollection(activeClient, args[0])
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(col, nil)
		},
	}
	cmd.ValidArgsFunction = completeCollectionIDs
	return cmd
}

func newCollectionsCreateCmd() *cobra.Command {
	var name, collectionType, title, shortTitle string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a collection",
		Example: `  # Create a list collection
  stqry collections create --name highlights --type list --title "Highlights"

  # Create a tour with a localised title
  stqry collections create --name city-tour --type tour --title "City Tour" --lang en

  # Override the short title (used in compact UI views)
  stqry collections create --name city-tour --type tour --title "Grand City Walking Tour" --short-title "City Tour"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCollectionType(collectionType); err != nil {
				return err
			}
			lang := flagLang
			if lang == "" {
				lang = "en"
			}
			fields := map[string]interface{}{
				"name": name,
				"type": collectionType,
			}
			if title != "" {
				fields["title"] = map[string]interface{}{lang: title}
			}
			// The API requires short_title; default it to title when omitted.
			effectiveShort := shortTitle
			if effectiveShort == "" {
				effectiveShort = title
			}
			if effectiveShort != "" {
				fields["short_title"] = map[string]interface{}{lang: effectiveShort}
			}

			col, err := api.CreateCollection(activeClient, fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(col, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Collection name (required)")
	cmd.Flags().StringVar(&collectionType, "type", "", fmt.Sprintf("Collection type (required; one of: %s)", strings.Join(validCollectionTypes, ", ")))
	cmd.Flags().StringVar(&title, "title", "", "Collection title")
	cmd.Flags().StringVar(&shortTitle, "short-title", "", "Collection short title (defaults to --title if omitted)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")

	return cmd
}

func newCollectionsUpdateCmd() *cobra.Command {
	var name, title, shortTitle, description string
	var coverImageID, coverImageGridID, coverImageWideID, logoID, previewID int
	var mapViewEnabled bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a collection",
		Example: `  # Rename a collection
  stqry collections update 42 --name new-name

  # Update the title in French
  stqry collections update 42 --title "Tour de ville" --lang fr

  # Set a cover image and description
  stqry collections update 42 --cover-image-media-item-id 123 --description "A walking tour..."`,
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
				case "title":
					fields["title"] = map[string]interface{}{lang: title}
				case "short-title":
					fields["short_title"] = map[string]interface{}{lang: shortTitle}
				case "description":
					fields["description"] = map[string]interface{}{lang: description}
				case "cover-image-media-item-id":
					fields["cover_image_media_item_id"] = coverImageID
				case "cover-image-grid-media-item-id":
					fields["cover_image_grid_media_item_id"] = coverImageGridID
				case "cover-image-wide-media-item-id":
					fields["cover_image_wide_media_item_id"] = coverImageWideID
				case "logo-media-item-id":
					fields["logo_media_item_id"] = logoID
				case "preview-media-item-id":
					fields["preview_media_item_id"] = previewID
				case "map-view-enabled":
					fields["map_view_enabled"] = mapViewEnabled
				}
			})

			col, err := api.UpdateCollection(activeClient, args[0], fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(col, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Collection name")
	cmd.Flags().StringVar(&title, "title", "", "Collection title")
	cmd.Flags().StringVar(&shortTitle, "short-title", "", "Collection short title")
	cmd.Flags().StringVar(&description, "description", "", "Collection description")
	cmd.Flags().IntVar(&coverImageID, "cover-image-media-item-id", 0, "Cover image media item ID")
	cmd.Flags().IntVar(&coverImageGridID, "cover-image-grid-media-item-id", 0, "Grid cover image media item ID")
	cmd.Flags().IntVar(&coverImageWideID, "cover-image-wide-media-item-id", 0, "Wide cover image media item ID")
	cmd.Flags().IntVar(&logoID, "logo-media-item-id", 0, "Logo media item ID")
	cmd.Flags().IntVar(&previewID, "preview-media-item-id", 0, "Preview image media item ID")
	cmd.Flags().BoolVar(&mapViewEnabled, "map-view-enabled", false, "Enable the map view for the collection (tour map)")
	cmd.ValidArgsFunction = completeCollectionIDs

	return cmd
}

func newCollectionsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a collection",
		Example: `  # Delete a collection
  stqry collections delete 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := api.DeleteCollection(activeClient, args[0]); err != nil {
				printer.PrintError(err)
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	}
	cmd.ValidArgsFunction = completeCollectionIDs
	return cmd
}

func newCollectionsItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "Manage collection items",
		Example: `  # List items in a collection
  stqry collections items list 42

  # Add a screen to a collection
  stqry collections items add 42 --item-type Screen --item-id 99`,
	}

	cmd.AddCommand(newCollectionsItemsListCmd())
	cmd.AddCommand(newCollectionsItemsAddCmd())
	cmd.AddCommand(newCollectionsItemsRemoveCmd())
	cmd.AddCommand(newCollectionsItemsReorderCmd())

	return cmd
}

func newCollectionsItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <collection-id>",
		Short: "List items in a collection",
		Example: `  # List all items in a collection
  stqry collections items list 42

  # Filter with built-in jq (no external jq needed)
  stqry collections items list 42 --jq '.[].item_id'

  # Pipe to external jq (alternative)
  stqry collections items list 42 --quiet | jq '.[].id'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items, meta, err := api.ListCollectionItems(activeClient, args[0], nil)
			if err != nil {
				printer.PrintError(err)
				return err
			}

			var m *output.Meta
			if meta != nil {
				m = &output.Meta{Page: meta.Page, PerPage: meta.PerPage, Total: meta.Count}
			}
			return printer.PrintList([]string{"id", "item_type", "item_id", "position"}, items, m)
		},
	}
	cmd.ValidArgsFunction = completeCollectionIDs
	return cmd
}

func newCollectionsItemsAddCmd() *cobra.Command {
	var itemType, itemID string

	cmd := &cobra.Command{
		Use:   "add <collection-id>",
		Short: "Add an item to a collection",
		Example: `  # Add a screen to a collection
  stqry collections items add 42 --item-type Screen --item-id 99

  # Add to a specific site
  stqry collections items add 42 --item-type Screen --item-id 99 --site mysite`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := map[string]interface{}{
				"item_type": itemType,
				"item_id":   itemID,
			}
			item, err := api.CreateCollectionItem(activeClient, args[0], fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&itemType, "item-type", "", "Item type (required)")
	cmd.Flags().StringVar(&itemID, "item-id", "", "Item ID (required)")
	cmd.MarkFlagRequired("item-type")
	cmd.MarkFlagRequired("item-id")
	cmd.ValidArgsFunction = completeCollectionIDs

	return cmd
}

func newCollectionsItemsRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <collection-id> <item-id>",
		Short: "Remove an item from a collection",
		Example: `  # Remove an item from a collection
  stqry collections items remove 42 99`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := api.DeleteCollectionItem(activeClient, args[0], args[1]); err != nil {
				printer.PrintError(err)
				return err
			}
			fmt.Println("Removed.")
			return nil
		},
	}
	cmd.ValidArgsFunction = completeCollectionIDs
	return cmd
}

func newCollectionsItemsReorderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reorder <collection-id> <item-id>...",
		Short: "Reorder items in a collection",
		Example: `  # Reorder items in a collection (IDs in desired order)
  stqry collections items reorder 42 99 100 101`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionID := args[0]
			itemIDs := args[1:]
			if err := api.ReorderCollectionItems(activeClient, collectionID, itemIDs); err != nil {
				printer.PrintError(err)
				return err
			}
			fmt.Println("Reordered.")
			return nil
		},
	}
	cmd.ValidArgsFunction = completeCollectionIDs
	return cmd
}
