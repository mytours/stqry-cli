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

func validateScreenType(t string) error {
	for _, v := range validScreenTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("invalid screen type %q (valid: %s)", t, strings.Join(validScreenTypes, ", "))
}

func newScreensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screens",
		Short: "Manage screens",
		Example: `  # List all screens
  stqry screens list

  # Create a story screen
  stqry screens create --name welcome --type story`,
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
  stqry screens list --site mysite --page 2 --per-page 25`,
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
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a screen by ID",
		Example: `  # Get a screen by ID
  stqry screens get 42

  # Get screen details as JSON
  stqry screens get 42 --json`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			screen, err := api.GetScreen(activeClient, args[0])
			if err != nil {
				return err
			}
			return printer.PrintOne(screen, nil)
		},
	}
}

// ── screens create ────────────────────────────────────────────────────────────

func newScreensCreateCmd() *cobra.Command {
	var name, screenType, title string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new screen",
		Example: `  # Create a story screen
  stqry screens create --name welcome --type story

  # Create a web screen with a localised title
  stqry screens create --name map-view --type web --title "Map View" --lang en`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateScreenType(screenType); err != nil {
				return err
			}

			fields := map[string]interface{}{
				"name": name,
				"type": screenType,
			}
			if title != "" {
				if flagLang != "" {
					fields["title"] = map[string]interface{}{flagLang: title}
				} else {
					fields["title"] = map[string]interface{}{"en": title}
				}
			}

			screen, err := api.CreateScreen(activeClient, fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(screen, nil)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Screen name (required)")
	cmd.Flags().StringVar(&screenType, "type", "", fmt.Sprintf("Screen type (required; one of: %s)", strings.Join(validScreenTypes, ", ")))
	cmd.Flags().StringVar(&title, "title", "", "Screen title")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")

	return cmd
}

// ── screens update ────────────────────────────────────────────────────────────

func newScreensUpdateCmd() *cobra.Command {
	var name, title string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a screen",
		Example: `  # Rename a screen
  stqry screens update 42 --name new-name

  # Update the title in English
  stqry screens update 42 --title "Welcome Screen" --lang en`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := map[string]interface{}{}
			cmd.Flags().Visit(func(f *flag.Flag) {
				switch f.Name {
				case "name":
					fields["name"] = name
				case "title":
					if flagLang != "" {
						fields["title"] = map[string]interface{}{flagLang: title}
					} else {
						fields["title"] = title
					}
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

	return cmd
}

// ── screens delete ────────────────────────────────────────────────────────────

func newScreensDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a screen",
		Example: `  # Delete a screen
  stqry screens delete 42`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return api.DeleteScreen(activeClient, args[0])
		},
	}
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
	cmd.AddCommand(newSectionSubItemCmd("badges", "badge_items", "badge_item"))
	cmd.AddCommand(newSectionSubItemCmd("links", "link_items", "link_item"))
	cmd.AddCommand(newSectionSubItemCmd("media", "media_items", "media_item"))
	cmd.AddCommand(newSectionSubItemCmd("prices", "price_items", "price_item"))
	cmd.AddCommand(newSectionSubItemCmd("social", "social_items", "social_item"))
	cmd.AddCommand(newSectionSubItemCmd("hours", "opening_time_items", "opening_time_item"))

	return cmd
}

// ── sections list ─────────────────────────────────────────────────────────────

func newSectionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <screen-id>",
		Short: "List story sections for a screen",
		Example: `  # List all sections for a screen
  stqry screens sections list 42`,
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
  stqry screens sections get 99 --screen-id 42`,
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
	var sectionType, title string

	cmd := &cobra.Command{
		Use:   "add <screen-id>",
		Short: "Add a story section to a screen",
		Example: `  # Add a text section to a screen
  stqry screens sections add 42 --type text

  # Add a titled gallery section in French
  stqry screens sections add 42 --type gallery --title "Galerie" --lang fr`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if sectionType == "" {
				return fmt.Errorf("--type is required")
			}

			fields := map[string]interface{}{
				"type": sectionType,
			}
			if title != "" {
				if flagLang != "" {
					fields["title"] = map[string]interface{}{flagLang: title}
				} else {
					fields["title"] = title
				}
			}

			section, err := api.CreateStorySection(activeClient, args[0], fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(section, nil)
		},
	}

	cmd.Flags().StringVar(&sectionType, "type", "", "Section type (required)")
	cmd.Flags().StringVar(&title, "title", "", "Section title")

	return cmd
}

// ── sections update ───────────────────────────────────────────────────────────

func newSectionsUpdateCmd() *cobra.Command {
	var screenID, title string

	cmd := &cobra.Command{
		Use:   "update <section-id>",
		Short: "Update a story section",
		Example: `  # Update a section's title
  stqry screens sections update 99 --screen-id 42 --title "About"

  # Update title in French
  stqry screens sections update 99 --screen-id 42 --title "À propos" --lang fr`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}

			fields := map[string]interface{}{}
			cmd.Flags().Visit(func(f *flag.Flag) {
				switch f.Name {
				case "title":
					if flagLang != "" {
						fields["title"] = map[string]interface{}{flagLang: title}
					} else {
						fields["title"] = title
					}
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
// cmdName is the CLI name, apiPath is the plural API segment, singularKey is the singular.
func newSectionSubItemCmd(cmdName, apiPath, singularKey string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: fmt.Sprintf("Manage %s sub-items", cmdName),
	}

	cmd.AddCommand(newSubItemListCmd(cmdName, apiPath))
	cmd.AddCommand(newSubItemAddCmd(cmdName, apiPath, singularKey))
	cmd.AddCommand(newSubItemUpdateCmd(cmdName, apiPath, singularKey))
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
			group: "  # List links in a section\n  stqry screens sections links list --screen-id 42 --section-id 99\n\n  # Add a web link to a section\n  stqry screens sections links add --screen-id 42 --section-id 99 --link-type web --url https://example.com --label \"Visit Site\"",
			list:  "  # List links in a section\n  stqry screens sections links list --screen-id 42 --section-id 99",
			add:   "  # Add a web link to a section\n  stqry screens sections links add --screen-id 42 --section-id 99 --link-type web --url https://example.com --label \"Visit Site\"",
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
			group: "  # List social items in a section\n  stqry screens sections social list --screen-id 42 --section-id 99\n\n  # Add a social link to a section\n  stqry screens sections social add --screen-id 42 --section-id 99 --social-network instagram --url https://instagram.com/example",
			list:  "  # List social items in a section\n  stqry screens sections social list --screen-id 42 --section-id 99",
			add:   "  # Add a social link to a section\n  stqry screens sections social add --screen-id 42 --section-id 99 --social-network instagram --url https://instagram.com/example",
		},
		"hours": {
			group: "  # List opening hours in a section\n  stqry screens sections hours list --screen-id 42 --section-id 99\n\n  # Add an opening hours entry\n  stqry screens sections hours add --screen-id 42 --section-id 99 --description \"Monday–Friday\" --time \"9:00–17:00\"",
			list:  "  # List opening hours in a section\n  stqry screens sections hours list --screen-id 42 --section-id 99",
			add:   "  # Add an opening hours entry\n  stqry screens sections hours add --screen-id 42 --section-id 99 --description \"Monday–Friday\" --time \"9:00–17:00\"",
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
func newSubItemAddCmd(cmdName, apiPath, singularKey string) *cobra.Command {
	var screenID, sectionID string

	cmd := &cobra.Command{
		Use:   "add",
		Short: fmt.Sprintf("Add a %s", singularKey),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			fields := map[string]interface{}{}
			cmd.Flags().Visit(func(f *flag.Flag) {
				// Skip meta flags
				if f.Name == "screen-id" || f.Name == "section-id" {
					return
				}
				fields[f.Name] = f.Value.String()
			})

			item, err := api.CreateSectionSubItem(activeClient, screenID, sectionID, apiPath, singularKey, fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")
	cmd.Flags().StringVar(&sectionID, "section-id", "", "Section ID (required)")

	// Type-specific flags
	switch cmdName {
	case "badges":
		cmd.Flags().String("badge-id", "", "Badge ID")
	case "links":
		cmd.Flags().String("link-type", "", "Link type")
		cmd.Flags().String("url", "", "URL")
		cmd.Flags().String("label", "", "Label")
	case "media":
		cmd.Flags().String("media-item-id", "", "Media item ID")
	case "prices":
		cmd.Flags().Int("price-cents", 0, "Price in cents")
		cmd.Flags().String("price-currency", "", "Price currency code")
		cmd.Flags().String("description", "", "Price description")
	case "social":
		cmd.Flags().String("social-network", "", "Social network name")
		cmd.Flags().String("url", "", "URL")
	case "hours":
		cmd.Flags().String("description", "", "Hours description")
		cmd.Flags().String("time", "", "Time string")
	}

	return cmd
}

// newSubItemUpdateCmd builds the update subcommand for a sub-item type.
func newSubItemUpdateCmd(cmdName, apiPath, singularKey string) *cobra.Command {
	var screenID, sectionID string

	cmd := &cobra.Command{
		Use:   "update <item-id>",
		Short: fmt.Sprintf("Update a %s", singularKey),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if screenID == "" {
				return fmt.Errorf("--screen-id is required")
			}
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			fields := map[string]interface{}{}
			cmd.Flags().Visit(func(f *flag.Flag) {
				if f.Name == "screen-id" || f.Name == "section-id" {
					return
				}
				fields[f.Name] = f.Value.String()
			})

			item, err := api.UpdateSectionSubItem(activeClient, screenID, sectionID, apiPath, args[0], singularKey, fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(item, nil)
		},
	}

	cmd.Flags().StringVar(&screenID, "screen-id", "", "Screen ID (required)")
	cmd.Flags().StringVar(&sectionID, "section-id", "", "Section ID (required)")

	return cmd
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
