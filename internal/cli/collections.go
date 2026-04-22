package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// hexColourPattern matches a CSS hex colour (3 or 6 digits, leading # optional).
// The server enforces "valid CSS hex color code"; passing e.g. `red` returns
// HTTP 422 "Map pin colour must be a valid CSS hex color code".
var hexColourPattern = regexp.MustCompile(`^#?[0-9A-Fa-f]{3}([0-9A-Fa-f]{3})?$`)

// validateMapPinColour accepts the literal "default" (reset sentinel) or a
// CSS hex colour. Empty string is a no-op for the Visit() pattern.
func validateMapPinColour(c string) error {
	if c == "" || c == "default" {
		return nil
	}
	if hexColourPattern.MatchString(c) {
		return nil
	}
	return fmt.Errorf("invalid map pin colour %q (must be a CSS hex code like \"#FF0000\" or \"FF0000\", or \"default\" to reset)", c)
}

// validCollectionTypes mirrors Collection::SUBTYPES_SHORT in mytours-web
// (app/models/collection.rb). Keep in sync if new subtypes are added.
var validCollectionTypes = []string{"list", "tour", "organization", "menu", "search"}

// validTourTypes mirrors the CollectionPartial.tour_type enum in
// docs/public_api.json. Set on `type: tour` collections to drive tour-specific
// iconography and copy in client apps (e.g. a walking icon for a city walk).
var validTourTypes = []string{
	"walking", "4wd", "aboriginal_site", "airplane", "bus", "canoeing",
	"cycling", "driving", "gallery", "helicopter", "historic_house",
	"horse_trail", "museum", "nature_trail", "scavenger_hunt", "ship", "train",
}

func validateCollectionType(t string) error {
	for _, v := range validCollectionTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid collection type %q (valid: %s)", t, strings.Join(validCollectionTypes, ", "))
}

func validateTourType(t string) error {
	if t == "" {
		return nil
	}
	for _, v := range validTourTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid tour type %q (valid: %s)", t, strings.Join(validTourTypes, ", "))
}

func newCollectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collections",
		Short: "Manage collections",
		Example: `  # List all collections
  stqry collections list

  # Create a tour collection
  stqry collections create --name "City Tour" --type tour`,
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
	var name, collectionType, title, shortTitle, description, tourType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a collection",
		Example: `  # Create a list collection
  stqry collections create --name "Highlights" --type list

  # Set the translatable title field (defaults to the primary language when --lang is omitted)
  stqry collections create --name "City Tour" --title "Tour de ville" --type tour --lang fr

  # Override the short title (used in compact UI views)
  stqry collections create --name "Grand City Walking Tour" --short-title "City Tour" --type tour

  # Create a tour with a description (saves a follow-up update call)
  stqry collections create --name "City Tour" --type tour --description "A walking tour of downtown"

  # Tag the tour's mode of transport so clients can show the right icon / copy
  stqry collections create --name "City Tour" --type tour --tour-type walking`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCollectionType(collectionType); err != nil {
				return err
			}
			if err := validateTourType(tourType); err != nil {
				return err
			}
			if name == "" && title == "" {
				return fmt.Errorf("either --name or --title is required")
			}
			lang := flagLang
			if lang == "" {
				lang = "en"
			}
			// Default name to title and title to name verbatim. Do not slugify;
			// the "name" field is just a flat-string display label, not a URL slug.
			effectiveName := name
			if effectiveName == "" {
				effectiveName = title
			}
			effectiveTitle := title
			if effectiveTitle == "" {
				effectiveTitle = name
			}
			fields := map[string]interface{}{
				"name": effectiveName,
				"type": collectionType,
			}
			fields["title"] = map[string]interface{}{lang: effectiveTitle}
			// The API requires short_title; default it to title when omitted.
			effectiveShort := shortTitle
			if effectiveShort == "" {
				effectiveShort = effectiveTitle
			}
			fields["short_title"] = map[string]interface{}{lang: effectiveShort}
			if description != "" {
				fields["description"] = map[string]interface{}{lang: description}
			}
			if tourType != "" {
				fields["tour_type"] = tourType
			}

			col, err := api.CreateCollection(activeClient, fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(col, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Collection name (defaults to --title if omitted; plain label, not a slug)")
	cmd.Flags().StringVar(&collectionType, "type", "", fmt.Sprintf("Collection type (required; one of: %s)", strings.Join(validCollectionTypes, ", ")))
	cmd.Flags().StringVar(&title, "title", "", "Collection title (defaults to --name if omitted)")
	cmd.Flags().StringVar(&shortTitle, "short-title", "", "Collection short title (defaults to --title if omitted)")
	cmd.Flags().StringVar(&description, "description", "", "Collection description")
	cmd.Flags().StringVar(&tourType, "tour-type", "", fmt.Sprintf("Tour mode of transport / venue (one of: %s)", strings.Join(validTourTypes, ", ")))
	cmd.MarkFlagRequired("type")

	return cmd
}

func newCollectionsUpdateCmd() *cobra.Command {
	var name, title, shortTitle, description, tourType string
	var coverImageID, coverImageGridID, coverImageWideID, logoID, previewID int
	var mapViewEnabled bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a collection",
		Example: `  # Rename a collection
  stqry collections update 42 --name "New Name"

  # Update the title in French
  stqry collections update 42 --title "Tour de ville" --lang fr

  # Set a cover image and description
  stqry collections update 42 --cover-image-media-item-id 123 --description "A walking tour..."

  # Tag the tour's mode of transport so clients can show the right icon / copy
  stqry collections update 42 --tour-type walking`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTourType(tourType); err != nil {
				return err
			}
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
				case "tour-type":
					fields["tour_type"] = tourType
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
	cmd.Flags().StringVar(&tourType, "tour-type", "", fmt.Sprintf("Tour mode of transport / venue (one of: %s)", strings.Join(validTourTypes, ", ")))
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
	cmd.AddCommand(newCollectionsItemsGetCmd())
	cmd.AddCommand(newCollectionsItemsAddCmd())
	cmd.AddCommand(newCollectionsItemsUpdateCmd())
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

func newCollectionsItemsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <collection-id> <item-id>",
		Short: "Get a single collection item",
		Example: `  # Get a collection item by ID
  stqry collections items get 42 99

  # Filter a specific field
  stqry collections items get 42 99 --jq '.position'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			item, err := api.GetCollectionItem(activeClient, args[0], args[1])
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}
	cmd.ValidArgsFunction = completeCollectionIDs
	return cmd
}

func newCollectionsItemsAddCmd() *cobra.Command {
	var itemType, itemID string
	var position int

	cmd := &cobra.Command{
		Use:   "add <collection-id>",
		Short: "Add an item to a collection",
		Example: `  # Add a screen to a collection (appended to the end)
  stqry collections items add 42 --item-type Screen --item-id 99

  # Insert at a specific position (0-based) so a follow-up reorder is not needed
  stqry collections items add 42 --item-type Screen --item-id 99 --position 0

  # Add to a specific site
  stqry collections items add 42 --item-type Screen --item-id 99 --site mysite`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := map[string]interface{}{
				"item_type": itemType,
				"item_id":   itemID,
			}
			if cmd.Flags().Changed("position") {
				fields["position"] = position
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
	cmd.Flags().IntVar(&position, "position", 0, "Position in the collection (0-based; omit to append to the end)")
	cmd.MarkFlagRequired("item-type")
	cmd.MarkFlagRequired("item-id")
	cmd.ValidArgsFunction = completeCollectionIDs

	return cmd
}

func newCollectionsItemsUpdateCmd() *cobra.Command {
	var position int
	var itemNumber string
	var lat, lng float64
	var mapPinIcon, mapPinStyle, mapPinColour, geofence string

	cmd := &cobra.Command{
		Use:   "update <collection-id> <item-id>",
		Short: "Update a collection item",
		Example: `  # Move a single item to a specific position (1-based)
  stqry collections items update 42 99 --position 3

  # Set GPS coordinates for a tour stop on the map
  stqry collections items update 42 99 --lat 42.9018 --lng -78.8728

  # Set a custom item number (e.g. "1A") shown in some UIs
  stqry collections items update 42 99 --item-number "1A"

  # Change the map pin colour (CSS hex, with or without leading #)
  stqry collections items update 42 99 --map-pin-colour "#FF6600"

  # Reset to the tour default pin colour
  stqry collections items update 42 99 --map-pin-colour default`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateMapPinColour(mapPinColour); err != nil {
				return err
			}
			fields := map[string]interface{}{}
			cmd.Flags().Visit(func(f *flag.Flag) {
				switch f.Name {
				case "position":
					fields["position"] = position
				case "item-number":
					fields["item_number"] = itemNumber
				case "lat":
					fields["lat"] = lat
				case "lng":
					fields["lng"] = lng
				case "map-pin-icon":
					fields["map_pin_icon"] = mapPinIcon
				case "map-pin-style":
					fields["map_pin_style"] = mapPinStyle
				case "map-pin-colour":
					fields["map_pin_colour"] = mapPinColour
				case "geofence":
					fields["geofence"] = geofence
				}
			})
			if len(fields) == 0 {
				return fmt.Errorf("no fields to update; pass at least one flag (e.g. --position)")
			}
			item, err := api.UpdateCollectionItem(activeClient, args[0], args[1], fields)
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().IntVar(&position, "position", 0, "Position in the collection (1-based)")
	cmd.Flags().StringVar(&itemNumber, "item-number", "", "Display number shown in list / map views")
	cmd.Flags().Float64Var(&lat, "lat", 0, "Latitude for the map pin")
	cmd.Flags().Float64Var(&lng, "lng", 0, "Longitude for the map pin")
	cmd.Flags().StringVar(&mapPinIcon, "map-pin-icon", "", "Map pin icon name. \"default\" resets to the tour default.")
	cmd.Flags().StringVar(&mapPinStyle, "map-pin-style", "", "Map pin style name. \"default\" resets to the tour default.")
	cmd.Flags().StringVar(&mapPinColour, "map-pin-colour", "", "Map pin colour as a CSS hex code (with or without leading #, e.g. \"#FF6600\" or \"FF6600\"), or \"default\" to reset. Validated client-side.")
	cmd.Flags().StringVar(&geofence, "geofence", "", "Geofence mode (e.g. off, on)")
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
