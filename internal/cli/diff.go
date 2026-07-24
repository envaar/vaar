/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		Short: "Compare dotenv key presence",
		Args:  exactDiffArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput && quiet {
				return NewToolError("--quiet and --json cannot be used together", nil)
			}

			leftPath := args[0]
			rightPath := args[1]

			leftData, err := readDiffFile(leftPath)
			if err != nil {
				return err
			}

			rightData, err := readDiffFile(rightPath)
			if err != nil {
				return err
			}

			result, err := diff.Compare(leftPath, leftData, rightPath, rightData)
			if err != nil {
				return NewToolError("comparing dotenv files", err)
			}

			different := result.HasDifferences()

			if jsonOutput {
				if err := writeDiffJSON(cmd, result, different); err != nil {
					return err
				}
			} else if !quiet {
				if err := writeDiffText(cmd, result); err != nil {
					return err
				}
			}

			if different {
				return ExitError{Code: ExitFindings}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Render key differences as JSON")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress normal output and use exit status only")

	return cmd
}

func exactDiffArgs(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return NewToolError("diff requires exactly two dotenv files", nil)
	}
	return nil
}

func readDiffFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	switch {
	case err == nil:
		if info.IsDir() {
			return nil, NewToolError(fmt.Sprintf("%s is a directory, expected a dotenv file", path), nil)
		}
		if !info.Mode().IsRegular() {
			return nil, NewToolError(fmt.Sprintf("%s is not a regular file, expected a dotenv file", path), nil)
		}
	case os.IsNotExist(err):
		return nil, NewToolError(fmt.Sprintf("reading %s: file does not exist", path), nil)
	default:
		return nil, NewToolError(fmt.Sprintf("reading %s", path), err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, NewToolError(fmt.Sprintf("reading %s", path), err)
	}
	return data, nil
}

func writeDiffText(cmd *cobra.Command, result diff.Result) error {
	if !result.HasDifferences() {
		return writeDiffLine(cmd, "No key differences found")
	}

	if len(result.MissingFromLeft) > 0 {
		if err := writeDiffLine(cmd, missingKeysLine(result.Left, result.MissingFromLeft)); err != nil {
			return err
		}
	}
	if len(result.MissingFromRight) > 0 {
		if err := writeDiffLine(cmd, missingKeysLine(result.Right, result.MissingFromRight)); err != nil {
			return err
		}
	}

	return nil
}

func writeDiffJSON(cmd *cobra.Command, result diff.Result, different bool) error {
	output := diffJSON{
		Left:             result.Left,
		Right:            result.Right,
		MissingFromLeft:  result.MissingFromLeft,
		MissingFromRight: result.MissingFromRight,
		Different:        different,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return NewToolError("rendering diff JSON output failed", err)
	}

	return writeDiffLine(cmd, string(data))
}

func writeDiffLine(cmd *cobra.Command, line string) error {
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
		return NewToolError("writing diff output failed", err)
	}
	return nil
}

func missingKeysLine(path string, keys []string) string {
	noun := "key"
	if len(keys) != 1 {
		noun = "keys"
	}
	return fmt.Sprintf("%s is missing %s: %s", path, noun, strings.Join(keys, ", "))
}
