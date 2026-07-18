package cli

import (
	"fmt"
	"os"

	"github.com/envaar/vaar/internal/diff"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Compare dotenv key presence",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return NewToolError("diff requires exactly two dotenv files", nil)
			}

			leftPath := args[0]
			rightPath := args[1]

			leftData, err := os.ReadFile(leftPath)
			if err != nil {
				return NewToolError(
					fmt.Sprintf("reading %s", leftPath),
					err,
				)
			}

			rightData, err := os.ReadFile(rightPath)
			if err != nil {
				return NewToolError(
					fmt.Sprintf("reading %s", rightPath),
					err,
				)
			}

			// === MODIFIED SECTION START ===
			res, err := diff.Compare(leftPath, leftData, rightPath, rightData)
			if err != nil {
				return NewToolError("comparing dotenv files", err)
			}

			// If both lists are empty, the files match structurally
			if len(res.MissingFromLeft) == 0 && len(res.MissingFromRight) == 0 {
				fmt.Println("Files are structurally identical.")
				return nil
			}

			// Helper to get correct singular/plural text
			pluralize := func(count int) string {
				if count == 1 {
					return "key"
				}
				return "keys"
			}

			// Print findings for the left file
			if len(res.MissingFromLeft) > 0 {
				fmt.Printf("%s is missing %d %s:\n", leftPath, len(res.MissingFromLeft), pluralize(len(res.MissingFromLeft)))
				for _, k := range res.MissingFromLeft {
					fmt.Printf("  - %s\n", k)
				}
			}

			// Print findings for the right file
			if len(res.MissingFromRight) > 0 {
				fmt.Printf("%s is missing %d %s:\n", rightPath, len(res.MissingFromRight), pluralize(len(res.MissingFromRight)))
				for _, k := range res.MissingFromRight {
					fmt.Printf("  - %s\n", k)
				}
			}

			// Return an error context to exit with a non-zero code since structural differences were found
			return NewToolError("dotenv files diverge", nil)
			// === MODIFIED SECTION END ===
		},
	}
}
