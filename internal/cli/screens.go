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

// validScreenTypes mirrors Screen::SUBTYPES_SHORT in mytours-web
// (app/models/screen.rb). Keep in sync if new subtypes are added server-side.
var validScreenTypes = []string{"story", "web", "panorama", "ar", "kiosk"}

// validHeaderLayouts mirrors the ScreenPartial.header_layout enum in
// docs/public_api.json. Controls what's rendered above the first section on a
// screen (e.g. a full-bleed cover image vs. no header at all).
var validHeaderLayouts = []string{"none", "image", "image_and_title", "short", "tall"}

// validLinkTypes mirrors StorySectionLinkItemPartial.link_type in
// docs/public_api.json. Drives icon / behaviour for each link inside a
// link_group section.
var validLinkTypes = []string{
	"twitter", "whatsapp", "wechat", "facebook", "instagram", "pinterest",
	"youtube", "vimeo", "linkedin", "tiktok", "weibo", "bluesky",
	"internal", "url", "email", "phone",
	"live_tours", "settings", "badges", "favourites", "download", "app_rating", "search",
}

// validIconTypes mirrors StorySectionLinkItemPartial.icon_type.
var validIconTypes = []string{"media_item", "stock_icon", "clear"}

// validLinkItemTypes mirrors StorySectionLinkItemPartial.item_type (the
// linked-resource type for `link_type: internal` links).
var validLinkItemTypes = []string{
	"Bundle", "Collection", "CollectionItem", "Screen", "MediaItem", "CrossRegionLink",
}

// validSocialNetworks mirrors StorySectionSocialItemPartial.social_network.
var validSocialNetworks = []string{
	"twitter", "whatsapp", "wechat", "facebook", "instagram", "pinterest",
	"youtube", "vimeo", "linkedin", "tiktok", "weibo", "bluesky",
}

// validSectionLayouts is the union of `layout` values accepted by story
// sections that have one: link_group (10 values) and social_group (icons,
// list). Other section types don't carry a layout field.
var validSectionLayouts = []string{
	"list", "button", "icon", "list_no_icon", "button_no_icon",
	"list_with_icon", "button_with_icon", "grid_image", "wide_image",
	"horizontal_slider", "icons",
}

// validSectionTypes mirrors the StorySection oneOf in docs/public_api.json.
// Keep in sync if new section schemas are added server-side.
var validSectionTypes = []string{
	"text",
	"single_media",
	"media_group",
	"image_slider",
	"link_group",
	"social_group",
	"location",
	"menu",
	"opening_time_group",
	"price_group",
	"badge_group",
	"quiz_question",
	"quiz_score",
	"form",
	"custom_widget",
}

func validateScreenType(t string) error {
	for _, v := range validScreenTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid screen type %q (valid: %s)", t, strings.Join(validScreenTypes, ", "))
}

func validateHeaderLayout(l string) error {
	return validateEnum("header layout", l, validHeaderLayouts)
}

func validateLinkType(t string) error      { return validateEnum("link type", t, validLinkTypes) }
func validateIconType(t string) error      { return validateEnum("icon type", t, validIconTypes) }
func validateLinkItemType(t string) error  { return validateEnum("item type", t, validLinkItemTypes) }
func validateSocialNetwork(n string) error { return validateEnum("social network", n, validSocialNetworks) }
func validateSectionLayout(l string) error { return validateEnum("section layout", l, validSectionLayouts) }

func validateEnum(label, value string, valid []string) error {
	if value == "" {
		return nil
	}
	for _, v := range valid {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf("invalid %s %q (valid: %s)", label, value, strings.Join(valid, ", "))
}

func validateSectionType(t string) error {
	for _, v := range validSectionTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid section type %q (valid: %s)", t, strings.Join(validSectionTypes, ", "))
}

func newScreensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screens",
		Short: "Manage screens",
		Example: `  # List all screens
  stqry screens list

  # Create a story screen
  stqry screens create --name "Welcome" --type story`,
	}

	cmd.AddCommand(newScreensListCmd())
	cmd.AddCommand(newScreensGetCmd())
	cmd.AddCommand(newScreensCreateCmd())
	cmd.AddCommand(newScreensUpdateCmd())
	cmd.AddCommand(newScreensDeleteCmd())
	cmd.AddCommand(newSectionsCmd())

	return cmd
}

// ── screens list ──────────────────────────────────────────────────────────────

func newScreensListCmd() *cobra.Command {
	var page, perPage int
	var q string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List screens",
		Example: `  # List all screens
  stqry screens list

  # Search for screens by name
  stqry screens list --q "welcome"

  # List using a specific site, paginated
  stqry screens list --site mysite --page 2 --per-page 25

  # Filter with built-in jq (no external jq needed)
  stqry screens list --jq '.[].name'

  # Pipe to external jq (alternative)
  stqry screens list --quiet | jq '.[].id'`,
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

			screens, meta, err := api.ListScreens(activeClient, query)
			if err != nil {
				return err
			}

			var outMeta *output.Meta
			if meta != nil {
				outMeta = &output.Meta{Page: meta.Page, PerPage: meta.PerPage, Total: meta.Count}
			}

			columns := []string{"id", "name", "title", "type"}
			return printer.PrintList(columns, screens, outMeta)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Results per page")
	cmd.Flags().StringVar(&q, "q", "", "Search query")

	return cmd
}

// ── screens get ───────────────────────────────────────────────────────────────

func newScreensGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a screen by ID",
		Example: `  # Get a screen by ID
  stqry screens get 42

  # Get screen details as JSON
  stqry screens get 42 --json

  # Filter a specific field
  stqry screens get 42 --jq '.name'`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			screen, err := api.GetScreen(activeClient, args[0])
			if err != nil {
				return err
			}
			return printer.PrintOne(screen, nil)
		},
	}
	cmd.ValidArgsFunction = completeScreenIDs
	return cmd
}

// ── screens create ────────────────────────────────────────────────────────────

func newScreensCreateCmd() *cobra.Command {
	var name, screenType, title, shortTitle, headerLayout string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new screen",
		Example: `  # Create a story screen
  stqry screens create --name "Welcome" --type story

  # Set the translatable title field (defaults to the primary language when --lang is omitted)
  stqry screens create --name "Map View" --title "Vue carte" --type web --lang fr

  # Override the short title (used in compact UI views)
  stqry screens create --name "Welcome to Our Tour" --short-title "Welcome" --type story

  # Pick a screen header layout up front
  stqry screens create --name "Stop 1 - Museum" --type story --header-layout image_and_title`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateScreenType(screenType); err != nil {
				return err
			}
			if err := validateHeaderLayout(headerLayout); err != nil {
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
				"type": screenType,
			}
			fields["title"] = map[string]interface{}{lang: effectiveTitle}
			// The API requires short_title; default it to title when omitted.
			effectiveShort := shortTitle
			if effectiveShort == "" {
				effectiveShort = effectiveTitle
			}
			fields["short_title"] = map[string]interface{}{lang: effectiveShort}
			if headerLayout != "" {
				fields["header_layout"] = headerLayout
			}

			screen, err := api.CreateScreen(activeClient, fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(screen, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Screen name (defaults to --title if omitted; plain label, not a slug)")
	cmd.Flags().StringVar(&screenType, "type", "", fmt.Sprintf("Screen type (required; one of: %s)", strings.Join(validScreenTypes, ", ")))
	cmd.Flags().StringVar(&title, "title", "", "Screen title (defaults to --name if omitted)")
	cmd.Flags().StringVar(&shortTitle, "short-title", "", "Screen short title (defaults to --title if omitted)")
	cmd.Flags().StringVar(&headerLayout, "header-layout", "", fmt.Sprintf("Header layout (one of: %s). Drives what's rendered above the first section; use instead of a redundant single_media section at the top.", strings.Join(validHeaderLayouts, ", ")))
	cmd.MarkFlagRequired("type")

	return cmd
}

// ── screens update ────────────────────────────────────────────────────────────

func newScreensUpdateCmd() *cobra.Command {
	var name, title, shortTitle, headerLayout string
	var coverImageID, coverImageGridID, coverImageWideID, backgroundImageID, logoID int

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a screen",
		Example: `  # Rename a screen
  stqry screens update 42 --name "New Name"

  # Update the title in English
  stqry screens update 42 --title "Welcome Screen" --lang en

  # Set a cover image
  stqry screens update 42 --cover-image-media-item-id 123

  # Promote the cover image to the screen header (instead of a single_media
  # section at the top of the screen)
  stqry screens update 42 --header-layout image_and_title`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateHeaderLayout(headerLayout); err != nil {
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
				case "header-layout":
					fields["header_layout"] = headerLayout
				case "cover-image-media-item-id":
					fields["cover_image_media_item_id"] = coverImageID
				case "cover-image-grid-media-item-id":
					fields["cover_image_grid_media_item_id"] = coverImageGridID
				case "cover-image-wide-media-item-id":
					fields["cover_image_wide_media_item_id"] = coverImageWideID
				case "background-image-media-item-id":
					fields["background_image_media_item_id"] = backgroundImageID
				case "logo-media-item-id":
					fields["logo_media_item_id"] = logoID
				}
			})

			screen, err := api.UpdateScreen(activeClient, args[0], fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(screen, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New screen name")
	cmd.Flags().StringVar(&title, "title", "", "New screen title")
	cmd.Flags().StringVar(&shortTitle, "short-title", "", "New screen short title")
	cmd.Flags().StringVar(&headerLayout, "header-layout", "", fmt.Sprintf("Header layout (one of: %s). Drives what's rendered above the first section; use instead of a redundant single_media section at the top.", strings.Join(validHeaderLayouts, ", ")))
	cmd.Flags().IntVar(&coverImageID, "cover-image-media-item-id", 0, "Cover image media item ID")
	cmd.Flags().IntVar(&coverImageGridID, "cover-image-grid-media-item-id", 0, "Grid cover image media item ID")
	cmd.Flags().IntVar(&coverImageWideID, "cover-image-wide-media-item-id", 0, "Wide cover image media item ID")
	cmd.Flags().IntVar(&backgroundImageID, "background-image-media-item-id", 0, "Background image media item ID")
	cmd.Flags().IntVar(&logoID, "logo-media-item-id", 0, "Logo media item ID")
	cmd.ValidArgsFunction = completeScreenIDs

	return cmd
}

// ── screens delete ────────────────────────────────────────────────────────────

func newScreensDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a screen",
		Example: `  # Delete a screen
  stqry screens delete 42`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return api.DeleteScreen(activeClient, args[0])
		},
	}
	cmd.ValidArgsFunction = completeScreenIDs
	return cmd
}

// ── screens sections ──────────────────────────────────────────────────────────

func newSectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sections",
		Short: "Manage story sections",
		Example: `  # List sections for a screen
  stqry screens sections list 42

  # Add a text section to a screen
  stqry screens sections add 42 --type text`,
	}

	cmd.AddCommand(newSectionsListCmd())
	cmd.AddCommand(newSectionsGetCmd())
	cmd.AddCommand(newSectionsAddCmd())
	cmd.AddCommand(newSectionsUpdateCmd())
	cmd.AddCommand(newSectionsRemoveCmd())
	cmd.AddCommand(newSectionsReorderCmd())

	// Sub-item type commands
	cmd.AddCommand(newSectionSubItemCmd("badges", "badge_items"))
	cmd.AddCommand(newSectionSubItemCmd("links", "link_items"))
	cmd.AddCommand(newSectionSubItemCmd("media", "media_items"))
	cmd.AddCommand(newSectionSubItemCmd("prices", "price_items"))
	cmd.AddCommand(newSectionSubItemCmd("social", "social_items"))
	cmd.AddCommand(newSectionSubItemCmd("hours", "opening_time_items"))

	return cmd
}

// ── sections list ─────────────────────────────────────────────────────────────

func newSectionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <screen-id>",
		Short: "List story sections for a screen",
		Example: `  # List all sections for a screen
  stqry screens sections list 42

  # Filter with built-in jq (no external jq needed)
  stqry screens sections list 42 --jq '.[].type'

  # Pipe to external jq (alternative)
  stqry screens sections list 42 --quiet | jq '.[].id'`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sections, meta, err := api.ListStorySections(activeClient, args[0], nil)
			if err != nil {
				return err
			}

			var outMeta *output.Meta
			if meta != nil {
				outMeta = &output.Meta{Page: meta.Page, PerPage: meta.PerPage, Total: meta.Count}
			}

			columns := []string{"id", "type", "position", "title"}
			return printer.PrintList(columns, sections, outMeta)
		},
	}
}

// ── sections get ──────────────────────────────────────────────────────────────

func newSectionsGetCmd() *cobra.Command {
	var screenID string

	cmd := &cobra.Command{
		Use:   "get <section-id>",
		Short: "Get a story section by ID",
		Example: `  # Get a section by ID
  stqry screens sections get 99 --screen-id 42

  # Filter a specific field
  stqry screens sections get 99 --screen-id 42 --jq '.type'`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			section, err := api.GetStorySection(activeClient, screenID, args[0])
			if err != nil {
				return err
			}
			return printer.PrintOne(section, nil)
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")

	return cmd
}

// ── sections add ──────────────────────────────────────────────────────────────

func newSectionsAddCmd() *cobra.Command {
	var sectionType, title, subtitle, body, description, textPosition, mapType, displayAddress string
	var mediaItemID int
	var lat, lng float64
	var directionsEnabled bool

	cmd := &cobra.Command{
		Use:   "add <screen-id>",
		Short: "Add a story section to a screen",
		Example: `  # Add a text section with body content
  stqry screens sections add 42 --type text --title "About" --body "Welcome to our tour."

  # Add a single_media section with an image
  stqry screens sections add 42 --type single_media --media-item-id 99 --description "A photo caption"

  # Add a titled gallery section in French
  stqry screens sections add 42 --type media_group --title "Galerie" --lang fr`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateSectionType(sectionType); err != nil {
				return err
			}

			lang := flagLang
			if lang == "" {
				lang = "en"
			}
			fields := map[string]interface{}{
				"type": sectionType,
			}
			if title != "" {
				fields["title"] = map[string]interface{}{lang: title}
			}
			if subtitle != "" {
				fields["subtitle"] = map[string]interface{}{lang: subtitle}
			}
			if body != "" {
				fields["body"] = map[string]interface{}{lang: body}
			}
			if description != "" {
				fields["description"] = map[string]interface{}{lang: description}
			}
			if mediaItemID != 0 {
				fields["media_item_id"] = mediaItemID
			}
			// single_media sections require text_position; default to "none" if not provided.
			if sectionType == "single_media" {
				if textPosition == "" {
					textPosition = "none"
				}
				fields["text_position"] = textPosition
			} else if textPosition != "" {
				fields["text_position"] = textPosition
			}
			// Location section fields. Visit() lets us distinguish "not passed" from "zero".
			cmd.Flags().Visit(func(f *flag.Flag) {
				switch f.Name {
				case "lat":
					fields["lat"] = lat
				case "lng":
					fields["lng"] = lng
				case "map-type":
					fields["map_type"] = mapType
				case "display-address":
					fields["display_address"] = map[string]interface{}{lang: displayAddress}
				case "directions-enabled":
					fields["directions_enabled"] = directionsEnabled
				}
			})
			// Location sections need a map_type; default to single_location if lat/lng given without one.
			if sectionType == "location" {
				if _, ok := fields["map_type"]; !ok {
					fields["map_type"] = "single_location"
				}
			}

			section, err := api.CreateStorySection(activeClient, args[0], fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(section, nil)
		},
	}

	cmd.Flags().StringVar(&sectionType, "type", "", fmt.Sprintf("Section type (required; one of: %s)", strings.Join(validSectionTypes, ", ")))
	cmd.MarkFlagRequired("type")
	cmd.Flags().StringVar(&title, "title", "", "Section title")
	cmd.Flags().StringVar(&subtitle, "subtitle", "", "Section subtitle (text sections only)")
	cmd.Flags().StringVar(&body, "body", "", "Section body (text sections only)")
	cmd.Flags().StringVar(&description, "description", "", "Section description (media sections only)")
	cmd.Flags().IntVar(&mediaItemID, "media-item-id", 0, "Media item ID (single_media sections only)")
	cmd.Flags().StringVar(&textPosition, "text-position", "", "Text position for single_media sections (left, right, top, bottom, none; default: none)")
	cmd.Flags().Float64Var(&lat, "lat", 0, "Latitude (location sections)")
	cmd.Flags().Float64Var(&lng, "lng", 0, "Longitude (location sections)")
	cmd.Flags().StringVar(&mapType, "map-type", "", "Map type for location sections (single_location, multiple_locations; default: single_location)")
	cmd.Flags().StringVar(&displayAddress, "display-address", "", "Display address for location sections")
	cmd.Flags().BoolVar(&directionsEnabled, "directions-enabled", false, "Enable directions for location sections")

	return cmd
}

// ── sections update ───────────────────────────────────────────────────────────

func newSectionsUpdateCmd() *cobra.Command {
	var screenID, title, subtitle, body, description, textPosition, layout string
	var mediaItemID int

	cmd := &cobra.Command{
		Use:   "update <section-id>",
		Short: "Update a story section",
		Example: `  # Update a section's title
  stqry screens sections update 99 --screen-id 42 --title "About"

  # Add body text to a text section
  stqry screens sections update 99 --screen-id 42 --body "Detailed description..."

  # Attach a media item to a single_media section
  stqry screens sections update 99 --screen-id 42 --media-item-id 123

  # Change a link_group section's layout from list to buttons (all links
  # inside the section render as buttons instead of list rows).
  stqry screens sections update 99 --screen-id 42 --layout button_with_icon`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			if err := validateSectionLayout(layout); err != nil {
				return err
			}

			lang := flagLang
			if lang == "" {
				lang = "en"
			}
			fields := map[string]interface{}{}
			cmd.Flags().Visit(func(f *flag.Flag) {
				switch f.Name {
				case "title":
					fields["title"] = map[string]interface{}{lang: title}
				case "subtitle":
					fields["subtitle"] = map[string]interface{}{lang: subtitle}
				case "body":
					fields["body"] = map[string]interface{}{lang: body}
				case "description":
					fields["description"] = map[string]interface{}{lang: description}
				case "media-item-id":
					fields["media_item_id"] = mediaItemID
				case "text-position":
					fields["text_position"] = textPosition
				case "layout":
					fields["layout"] = layout
				}
			})

			section, err := api.UpdateStorySection(activeClient, screenID, args[0], fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(section, nil)
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")
	cmd.Flags().StringVar(&title, "title", "", "New section title")
	cmd.Flags().StringVar(&subtitle, "subtitle", "", "New section subtitle (text sections only)")
	cmd.Flags().StringVar(&body, "body", "", "New section body (text sections only)")
	cmd.Flags().StringVar(&description, "description", "", "New section description (media sections only)")
	cmd.Flags().IntVar(&mediaItemID, "media-item-id", 0, "New media item ID (single_media sections only)")
	cmd.Flags().StringVar(&textPosition, "text-position", "", "Text position for single_media sections (left, right, top, bottom, none)")
	cmd.Flags().StringVar(&layout, "layout", "", fmt.Sprintf("Section layout. link_group: list, button, icon, list_no_icon, button_no_icon, list_with_icon, button_with_icon, grid_image, wide_image, horizontal_slider. social_group: icons, list. (union validated: %s)", strings.Join(validSectionLayouts, ", ")))

	return cmd
}

// ── sections remove ───────────────────────────────────────────────────────────

func newSectionsRemoveCmd() *cobra.Command {
	var screenID string

	cmd := &cobra.Command{
		Use:   "remove <section-id>",
		Short: "Remove a story section",
		Example: `  # Remove a section from a screen
  stqry screens sections remove 99 --screen-id 42`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			return api.DeleteStorySection(activeClient, screenID, args[0])
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")

	return cmd
}

// ── sections reorder ──────────────────────────────────────────────────────────

func newSectionsReorderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reorder <screen-id> <section-id>...",
		Short: "Reorder story sections",
		Example: `  # Reorder sections on a screen (specify section IDs in desired order)
  stqry screens sections reorder 42 99 100 101`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			screenID := args[0]
			sectionIDs := args[1:]
			return api.ReorderStorySections(activeClient, screenID, sectionIDs)
		},
	}
}

// ── generic sub-item command factory ─────────────────────────────────────────

// newSectionSubItemCmd creates a command group for a sub-item type (e.g. badges).
// cmdName is the CLI name, apiPath is the plural API segment.
func newSectionSubItemCmd(cmdName, apiPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: fmt.Sprintf("Manage %s sub-items", cmdName),
	}

	cmd.AddCommand(newSubItemListCmd(cmdName, apiPath))
	cmd.AddCommand(newSubItemAddCmd(cmdName, apiPath))
	cmd.AddCommand(newSubItemUpdateCmd(cmdName, apiPath))
	cmd.AddCommand(newSubItemRemoveCmd(cmdName, apiPath))

	type subItemExamples struct {
		group string
		list  string
		add   string
	}
	exampleMap := map[string]subItemExamples{
		"badges": {
			group: "  # List badges in a section\n  stqry screens sections badges list --screen-id 42 --section-id 99\n\n  # Add a badge to a section\n  stqry screens sections badges add --screen-id 42 --section-id 99 --badge-id 7",
			list:  "  # List badges in a section\n  stqry screens sections badges list --screen-id 42 --section-id 99",
			add:   "  # Add a badge to a section\n  stqry screens sections badges add --screen-id 42 --section-id 99 --badge-id 7",
		},
		"links": {
			group: "  # List links in a section\n  stqry screens sections links list --screen-id 42 --section-id 99\n\n  # Add a URL link (icon is auto-chosen from --link-type; url gets a globe)\n  stqry screens sections links add --screen-id 42 --section-id 99 \\\n    --link-type url --url https://example.com --link-text \"Visit site\"\n\n  # Add a social link (handle in --username; icon is the network's logo)\n  stqry screens sections links add --screen-id 42 --section-id 99 \\\n    --link-type instagram --username @example --link-text \"Follow on Instagram\"\n\n  # Add an internal link to another screen\n  stqry screens sections links add --screen-id 42 --section-id 99 \\\n    --link-type internal --item-type Screen --item-id 100 --link-text \"Related stop\"",
			list:  "  # List links in a section\n  stqry screens sections links list --screen-id 42 --section-id 99",
			add:   "  # Add a URL link (icon is auto-chosen from --link-type)\n  stqry screens sections links add --screen-id 42 --section-id 99 \\\n    --link-type url --url https://example.com --link-text \"Visit site\"",
		},
		"media": {
			group: "  # List media items in a section\n  stqry screens sections media list --screen-id 42 --section-id 99\n\n  # Add a media item to a section\n  stqry screens sections media add --screen-id 42 --section-id 99 --media-item-id 55",
			list:  "  # List media items in a section\n  stqry screens sections media list --screen-id 42 --section-id 99",
			add:   "  # Add a media item to a section\n  stqry screens sections media add --screen-id 42 --section-id 99 --media-item-id 55",
		},
		"prices": {
			group: "  # List price items in a section\n  stqry screens sections prices list --screen-id 42 --section-id 99\n\n  # Add a price to a section\n  stqry screens sections prices add --screen-id 42 --section-id 99 --price-cents 1500 --price-currency USD --description \"Adult\"",
			list:  "  # List price items in a section\n  stqry screens sections prices list --screen-id 42 --section-id 99",
			add:   "  # Add a price to a section\n  stqry screens sections prices add --screen-id 42 --section-id 99 --price-cents 1500 --price-currency USD --description \"Adult\"",
		},
		"social": {
			group: "  # List social items in a section\n  stqry screens sections social list --screen-id 42 --section-id 99\n\n  # Add a social handle to a social_group section\n  stqry screens sections social add --screen-id 42 --section-id 99 \\\n    --social-network instagram --username @example --link-text \"Follow us\"",
			list:  "  # List social items in a section\n  stqry screens sections social list --screen-id 42 --section-id 99",
			add:   "  # Add a social handle to a social_group section\n  stqry screens sections social add --screen-id 42 --section-id 99 \\\n    --social-network instagram --username @example --link-text \"Follow us\"",
		},
		"hours": {
			group: "  # List opening hours in a section\n  stqry screens sections hours list --screen-id 42 --section-id 99\n\n  # Add an opening hours entry\n  stqry screens sections hours add --screen-id 42 --section-id 99 --description \"Monday-Friday\" --time \"9:00-17:00\"",
			list:  "  # List opening hours in a section\n  stqry screens sections hours list --screen-id 42 --section-id 99",
			add:   "  # Add an opening hours entry\n  stqry screens sections hours add --screen-id 42 --section-id 99 --description \"Monday-Friday\" --time \"9:00-17:00\"",
		},
	}
	if ex, ok := exampleMap[cmdName]; ok {
		cmd.Example = ex.group
		for _, sub := range cmd.Commands() {
			switch sub.Name() {
			case "list":
				sub.Example = ex.list
			case "add":
				sub.Example = ex.add
			}
		}
	}

	return cmd
}

// newSubItemListCmd builds the list subcommand for a sub-item type.
func newSubItemListCmd(cmdName, apiPath string) *cobra.Command {
	var screenID, sectionID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s", cmdName),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}
			items, err := api.ListSectionSubItems(activeClient, screenID, sectionID, apiPath)
			if err != nil {
				return err
			}
			return printer.PrintList(nil, items, nil)
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")
	cmd.Flags().StringVar(&sectionID, "section-id", "", "Section ID (required)")

	return cmd
}

// newSubItemAddCmd builds the add subcommand for a sub-item type.
func newSubItemAddCmd(cmdName, apiPath string) *cobra.Command {
	var screenID, sectionID string

	cmd := &cobra.Command{
		Use:   "add",
		Short: fmt.Sprintf("Add a %s item", cmdName),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			lang := flagLang
			if lang == "" {
				lang = "en"
			}
			fields, err := buildSubItemFields(cmd, cmdName, lang)
			if err != nil {
				return err
			}

			item, err := api.CreateSectionSubItem(activeClient, screenID, sectionID, apiPath, fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")
	cmd.Flags().StringVar(&sectionID, "section-id", "", "Section ID (required)")
	registerSubItemFlags(cmd, cmdName)
	return cmd
}

// newSubItemUpdateCmd builds the update subcommand for a sub-item type.
func newSubItemUpdateCmd(cmdName, apiPath string) *cobra.Command {
	var screenID, sectionID string

	cmd := &cobra.Command{
		Use:   "update <item-id>",
		Short: fmt.Sprintf("Update a %s item", cmdName),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			lang := flagLang
			if lang == "" {
				lang = "en"
			}
			fields, err := buildSubItemFields(cmd, cmdName, lang)
			if err != nil {
				return err
			}

			item, err := api.UpdateSectionSubItem(activeClient, screenID, sectionID, apiPath, args[0], fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")
	cmd.Flags().StringVar(&sectionID, "section-id", "", "Section ID (required)")
	registerSubItemFlags(cmd, cmdName)
	return cmd
}

// registerSubItemFlags adds the type-specific flags for a sub-item command.
// Shared between `add` and `update` so the flag surface is symmetric.
func registerSubItemFlags(cmd *cobra.Command, cmdName string) {
	switch cmdName {
	case "badges":
		cmd.Flags().String("badge-id", "", "Badge ID")
	case "links":
		cmd.Flags().String("link-type", "", fmt.Sprintf("Link type (one of: %s)", strings.Join(validLinkTypes, ", ")))
		cmd.Flags().String("url", "", "Destination URL (TranslatedString; for link-type=url / phone / email / social)")
		cmd.Flags().String("link-text", "", "Display label for the link (TranslatedString)")
		cmd.Flags().String("username", "", "Username / handle (TranslatedString; used by social link types)")
		cmd.Flags().String("icon-type", "", fmt.Sprintf("Icon source (one of: %s). stock_icon is the default — the icon is auto-chosen from --link-type (globe for url, the network's logo for social types, etc.); you can't pick among stock icons. Pass clear to hide the icon, media_item to use an uploaded media item (Builder-only; no public endpoint to set the media item id from this CLI).", strings.Join(validIconTypes, ", ")))
		cmd.Flags().String("item-type", "", fmt.Sprintf("Linked-resource type for link-type=internal (one of: %s)", strings.Join(validLinkItemTypes, ", ")))
		cmd.Flags().Int("item-id", 0, "Linked-resource ID for link-type=internal")
		cmd.Flags().Int("position", 0, "Position within the section (0-based)")
	case "media":
		cmd.Flags().String("media-item-id", "", "Media item ID")
	case "prices":
		cmd.Flags().Int("price-cents", 0, "Price in cents")
		cmd.Flags().String("price-currency", "", "Price currency code")
		cmd.Flags().String("description", "", "Price description")
	case "social":
		cmd.Flags().String("social-network", "", fmt.Sprintf("Social network (one of: %s)", strings.Join(validSocialNetworks, ", ")))
		cmd.Flags().String("username", "", "Username / handle (TranslatedString)")
		cmd.Flags().String("link-text", "", "Display label (TranslatedString)")
		cmd.Flags().Int("position", 0, "Position within the section (0-based)")
	case "hours":
		cmd.Flags().String("description", "", "Hours description")
		cmd.Flags().String("time", "", "Time string")
	}
}

// buildSubItemFields extracts the set flags from cmd and maps them to the
// correct API field names. For fields typed as TranslatedString in the public
// API the flag value is wrapped in a {lang: value} map. For fields with an
// enum the value is validated client-side so typos surface before hitting the
// wire.
func buildSubItemFields(cmd *cobra.Command, cmdName, lang string) (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	var enumErr error
	cmd.Flags().Visit(func(f *flag.Flag) {
		if enumErr != nil {
			return
		}
		if f.Name == "screen-id" || f.Name == "section-id" {
			return
		}
		v := f.Value.String()
		switch cmdName {
		case "links":
			switch f.Name {
			case "link-type":
				if err := validateLinkType(v); err != nil {
					enumErr = err
					return
				}
				fields["link_type"] = v
			case "icon-type":
				if err := validateIconType(v); err != nil {
					enumErr = err
					return
				}
				fields["icon_type"] = v
			case "item-type":
				if err := validateLinkItemType(v); err != nil {
					enumErr = err
					return
				}
				fields["item_type"] = v
			case "url":
				fields["url"] = map[string]interface{}{lang: v}
			case "link-text":
				fields["link_text"] = map[string]interface{}{lang: v}
			case "username":
				fields["username"] = map[string]interface{}{lang: v}
			case "item-id":
				if n, err := strconv.Atoi(v); err == nil {
					fields["item_id"] = n
				}
			case "position":
				if n, err := strconv.Atoi(v); err == nil {
					fields["position"] = n
				}
			}
		case "social":
			switch f.Name {
			case "social-network":
				if err := validateSocialNetwork(v); err != nil {
					enumErr = err
					return
				}
				fields["social_network"] = v
			case "username":
				fields["username"] = map[string]interface{}{lang: v}
			case "link-text":
				fields["link_text"] = map[string]interface{}{lang: v}
			case "position":
				if n, err := strconv.Atoi(v); err == nil {
					fields["position"] = n
				}
			}
		default:
			// Generic path for types whose flag names already match the API
			// field names (badges, media, prices, hours).
			fields[f.Name] = v
		}
	})
	return fields, enumErr
}

// newSubItemRemoveCmd builds the remove subcommand for a sub-item type.
func newSubItemRemoveCmd(cmdName, apiPath string) *cobra.Command {
	var screenID, sectionID string

	cmd := &cobra.Command{
		Use:   "remove <item-id>",
		Short: fmt.Sprintf("Remove a %s item", cmdName),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}
			return api.DeleteSectionSubItem(activeClient, screenID, sectionID, apiPath, args[0])
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")
	cmd.Flags().StringVar(&sectionID, "section-id", "", "Section ID (required)")

	return cmd
}
