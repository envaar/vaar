/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/envaar/vaar/internal/diff"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <left> <right>",
		Short: "Compare dotenv key presence",
		Args:  exactDiffArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			res, err := diff.Compare(leftPath, leftData, rightPath, rightData)
			if err != nil {
				return NewToolError("comparing dotenv files", err)
			}

			if len(res.MissingFromLeft) == 0 && len(res.MissingFromRight) == 0 {
				if err := writeDiffLine(cmd, "No key differences found"); err != nil {
					return err
				}
				return nil
			}

			if len(res.MissingFromLeft) > 0 {
				if err := writeDiffLine(cmd, missingKeysLine(leftPath, res.MissingFromLeft)); err != nil {
					return err
				}
			}
			if len(res.MissingFromRight) > 0 {
				if err := writeDiffLine(cmd, missingKeysLine(rightPath, res.MissingFromRight)); err != nil {
					return err
				}
			}

			return ExitError{Code: ExitFindings}
		},
	}
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
