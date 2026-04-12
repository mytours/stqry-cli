package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
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
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, err := api.GetCollection(activeClient, args[0])
			if err != nil {
				printer.PrintError(err)
				return err
			}
			return printer.PrintOne(col, nil)
		},
	}
}

func newCollectionsCreateCmd() *cobra.Command {
	var name, collectionType, title string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a collection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCollectionType(collectionType); err != nil {
				return err
			}
			fields := map[string]interface{}{
				"name": name,
				"type": collectionType,
			}
			if title != "" {
				if flagLang != "" {
					fields["title"] = map[string]interface{}{flagLang: title}
				} else {
					fields["title"] = map[string]interface{}{"en": title}
				}
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
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")

	return cmd
}

func newCollectionsUpdateCmd() *cobra.Command {
	var name, title string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := map[string]interface{}{}
			if name != "" {
				fields["name"] = name
			}
			if title != "" {
				if flagLang != "" {
					fields["title"] = map[string]interface{}{flagLang: title}
				} else {
					fields["title"] = map[string]interface{}{"en": title}
				}
			}

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

	return cmd
}

func newCollectionsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := api.DeleteCollection(activeClient, args[0]); err != nil {
				printer.PrintError(err)
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	}
}

func newCollectionsItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "Manage collection items",
	}

	cmd.AddCommand(newCollectionsItemsListCmd())
	cmd.AddCommand(newCollectionsItemsAddCmd())
	cmd.AddCommand(newCollectionsItemsRemoveCmd())
	cmd.AddCommand(newCollectionsItemsReorderCmd())

	return cmd
}

func newCollectionsItemsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <collection-id>",
		Short: "List items in a collection",
		Args:  cobra.ExactArgs(1),
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
}

func newCollectionsItemsAddCmd() *cobra.Command {
	var itemType, itemID string

	cmd := &cobra.Command{
		Use:   "add <collection-id>",
		Short: "Add an item to a collection",
		Args:  cobra.ExactArgs(1),
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

	return cmd
}

func newCollectionsItemsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <collection-id> <item-id>",
		Short: "Remove an item from a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := api.DeleteCollectionItem(activeClient, args[0], args[1]); err != nil {
				printer.PrintError(err)
				return err
			}
			fmt.Println("Removed.")
			return nil
		},
	}
}

func newCollectionsItemsReorderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reorder <collection-id> <item-id>...",
		Short: "Reorder items in a collection",
		Args:  cobra.MinimumNArgs(2),
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
}
