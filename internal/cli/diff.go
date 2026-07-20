/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/envaar/vaar/internal/diff"
	"github.com/spf13/cobra"
)

type diffJSON struct {
	Left             string   `json:"left"`
	Right            string   `json:"right"`
	MissingFromLeft  []string `json:"missing_from_left"`
	MissingFromRight []string `json:"missing_from_right"`
	Different        bool     `json:"different"`
}

func newDiffCmd() *cobra.Command {
	var jsonOutput bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "diff <left> <right>",
		Short: "Compare dotenv key presence between two files",
		Args:  cobra.ExactArgs(2),

		RunE: func(cmd *cobra.Command, args []string) error {

			if jsonOutput && quiet {
				return fmt.Errorf("--quiet and --json cannot be used together")
			}

			leftData, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			rightData, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			result, err := diff.Compare(
				args[0],
				leftData,
				args[1],
				rightData,
			)

			if err != nil {
				return err
			}

			different := len(result.MissingFromLeft) > 0 ||
				len(result.MissingFromRight) > 0

			if jsonOutput {
				output := diffJSON{
					Left:             result.Left,
					Right:            result.Right,
					MissingFromLeft:  result.MissingFromLeft,
					MissingFromRight: result.MissingFromRight,
					Different:        different,
				}

				data, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return err
				}

				if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(data)); err != nil {
					return err
				}
			}

			if different {
				return ExitError{Code: ExitFindings}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(
		&jsonOutput,
		"json",
		false,
		"output machine-readable JSON",
	)

	cmd.Flags().BoolVar(
		&quiet,
		"quiet",
		false,
		"only return exit status",
	)

	return cmd
}
