package cli

import (
	"fmt"
	"strconv"

	"github.com/mytours/stqry-cli/internal/api"
	"github.com/mytours/stqry-cli/internal/output"
	"github.com/spf13/cobra"
)

func newCodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codes",
		Short: "Manage redemption codes",
		Example: `  # List all redemption codes
  stqry codes list

  # Create a redemption code
  stqry codes create --coupon-code WELCOME10 --linked-type Collection --linked-id 42 --project-id 1`,
	}

	cmd.AddCommand(newCodesListCmd())
	cmd.AddCommand(newCodesGetCmd())
	cmd.AddCommand(newCodesCreateCmd())
	cmd.AddCommand(newCodesUpdateCmd())
	cmd.AddCommand(newCodesDeleteCmd())

	return cmd
}

func newCodesListCmd() *cobra.Command {
	var page, perPage int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all codes",
		Example: `  # List all redemption codes
  stqry codes list

  # List using a specific site
  stqry codes list --site mysite`,
		RunE: func(cmd *cobra.Command, args []string) error {
			query := map[string]string{}
			if page > 0 {
				query["page"] = strconv.Itoa(page)
			}
			if perPage > 0 {
				query["per_page"] = strconv.Itoa(perPage)
			}

			codes, meta, err := api.ListCodes(activeClient, query)
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

			columns := []string{"id", "coupon_code", "linked_type", "linked_id", "status"}
			return printer.PrintList(columns, codes, outMeta)
		},
	}

	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 0, "Results per page")

	return cmd
}

func newCodesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a code by ID",
		Example: `  # Get a code by ID
  stqry codes get 10

  # Get code details as JSON
  stqry codes get 10 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := api.GetCode(activeClient, args[0])
			if err != nil {
				return err
			}
			return printer.PrintOne(code, &output.Meta{})
		},
	}
}

func newCodesCreateCmd() *cobra.Command {
	var couponCode, linkedType, validFrom, validTo string
	var linkedID, projectID, maxRedemptions int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new code",
		Example: `  # Create a basic redemption code
  stqry codes create --coupon-code WELCOME10 --linked-type Collection --linked-id 42 --project-id 1

  # Create a code with validity dates and a redemption limit
  stqry codes create --coupon-code SUMMER25 --linked-type Collection --linked-id 42 --project-id 1 --valid-from 2025-06-01 --valid-to 2025-08-31 --max-redemptions 100`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if couponCode == "" {
				return fmt.Errorf("--coupon-code is required")
			}
			if linkedType == "" {
				return fmt.Errorf("--linked-type is required")
			}
			if linkedID == 0 {
				return fmt.Errorf("--linked-id is required")
			}
			if projectID == 0 {
				return fmt.Errorf("--project-id is required")
			}

			fields := map[string]interface{}{
				"coupon_code": couponCode,
				"linked_type": linkedType,
				"linked_id":   linkedID,
				"project_id":  projectID,
			}
			if validFrom != "" {
				fields["valid_from"] = validFrom
			}
			if validTo != "" {
				fields["valid_to"] = validTo
			}
			if maxRedemptions > 0 {
				fields["max_redemptions"] = maxRedemptions
			}

			code, err := api.CreateCode(activeClient, fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(code, &output.Meta{})
		},
	}

	cmd.Flags().StringVar(&couponCode, "coupon-code", "", "Coupon code value (required)")
	cmd.Flags().StringVar(&linkedType, "linked-type", "", "Linked resource type (required)")
	cmd.Flags().IntVar(&linkedID, "linked-id", 0, "Linked resource ID (required)")
	cmd.Flags().IntVar(&projectID, "project-id", 0, "Project ID (required)")
	cmd.Flags().StringVar(&validFrom, "valid-from", "", "Valid from date")
	cmd.Flags().StringVar(&validTo, "valid-to", "", "Valid to date")
	cmd.Flags().IntVar(&maxRedemptions, "max-redemptions", 0, "Maximum number of redemptions")

	return cmd
}

func newCodesUpdateCmd() *cobra.Command {
	var couponCode, validFrom, validTo string
	var maxRedemptions int

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing code",
		Example: `  # Update the coupon code value
  stqry codes update 10 --coupon-code NEWCODE

  # Set an expiry date and a redemption cap
  stqry codes update 10 --valid-to 2025-12-31 --max-redemptions 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields := map[string]interface{}{}

			if cmd.Flags().Changed("coupon-code") {
				fields["coupon_code"] = couponCode
			}
			if cmd.Flags().Changed("valid-from") {
				fields["valid_from"] = validFrom
			}
			if cmd.Flags().Changed("valid-to") {
				fields["valid_to"] = validTo
			}
			if cmd.Flags().Changed("max-redemptions") {
				fields["max_redemptions"] = maxRedemptions
			}

			if len(fields) == 0 {
				return fmt.Errorf("no fields specified to update")
			}

			code, err := api.UpdateCode(activeClient, args[0], fields)
			if err != nil {
				return err
			}
			return printer.PrintOne(code, &output.Meta{})
		},
	}

	cmd.Flags().StringVar(&couponCode, "coupon-code", "", "Coupon code value")
	cmd.Flags().StringVar(&validFrom, "valid-from", "", "Valid from date")
	cmd.Flags().StringVar(&validTo, "valid-to", "", "Valid to date")
	cmd.Flags().IntVar(&maxRedemptions, "max-redemptions", 0, "Maximum number of redemptions")

	return cmd
}

func newCodesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a code by ID",
		Example: `  # Delete a redemption code
  stqry codes delete 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := api.DeleteCode(activeClient, args[0]); err != nil {
				return err
			}
			fmt.Println("Deleted.")
			return nil
		},
	}
}
