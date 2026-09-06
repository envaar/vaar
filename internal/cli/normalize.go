/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"

	"github.com/envaar/vaar/internal/lint"
	"github.com/spf13/cobra"
)

// newNormalizeCmd builds the normalize subcommand, which applies the blunt
// whole-file formatting normalization on demand, independent of any lint
// finding.
func newNormalizeCmd() *cobra.Command {
	var scope lint.Options

	cmd := &cobra.Command{
		Use:   "normalize",
		Short: "Normalize dotenv formatting on demand",
		Long: `Apply Vaar's blunt formatting normalization to discovered dotenv files,
whether or not any lint rule reports a finding.

It strips a leading UTF-8 BOM, converts every line ending to LF (including
uniform CRLF or lone-CR files, which lint --fix deliberately leaves alone),
trims trailing whitespace, collapses repeated blank lines to one and leaves
exactly one final newline. Each rewritten file is reported as it is written;
files that are already normalized are left untouched.

Use --target to normalize one file and --target-dir to discover files under
one directory. Use either --target or --target-dir, not both. To repair only
what the lint rules actually report, use vaar lint --fix instead.`,
		Example: `  vaar normalize
  vaar normalize --target=.env.staging
  vaar normalize --target-dir=src`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			normalized, err := lint.Normalize(lint.Options{
				Root:      ".",
				Target:    scope.Target,
				TargetDir: scope.TargetDir,
			})
			if err != nil {
				return NewToolError("normalize failed", err)
			}

			for _, path := range normalized {
				fmt.Fprintf(cmd.OutOrStdout(), "normalized %s\n", path)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&scope.Target, "target", "", "Normalize only the specified file path.")
	flags.StringVar(&scope.TargetDir, "target-dir", "", "Recursively normalize dotenv files under the specified directory.")

	return cmd
}
