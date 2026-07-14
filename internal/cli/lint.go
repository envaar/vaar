/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
	"github.com/envaar/vaar/internal/report"
	"github.com/spf13/cobra"
)

// newLintCmd builds the lint subcommand, wires the built-in rule set and
// exposes the flag surface for safe fixes, JSON output and rule selection.
func newLintCmd() *cobra.Command {
	var selection lint.Options
	var lintFix bool
	var lintJSON bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Run environment lint checks",
		Long: `Run Vaar's lint rules against discovered dotenv files in the current repository.

The command supports safe fixes, JSON output and repeatable rule selection
flags. Use --only to narrow the selected rules and --skip to remove rules after
selection:

  vaar lint --only=duplicate-key
  vaar lint --only=duplicate-key --only=invalid-key-name
  vaar lint --skip=trailing-whitespace
  vaar lint --skip=trailing-whitespace --skip=extra-blank-line`,
		Example: `  vaar lint
  vaar lint --json
  vaar lint --fix
  vaar lint --only=duplicate-key --only=invalid-key-name
  vaar lint --skip=trailing-whitespace --skip=extra-blank-line`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runner := lint.NewRunner(rules.All()...)
			result, err := runner.Run(cmd.Context(), lint.Options{
				Root:      ".",
				OnlyRules: selection.OnlyRules,
				SkipRules: selection.SkipRules,
				Fix:       lintFix,
			})
			if err != nil {
				return NewToolError("lint failed", err)
			}

			switch {
			case lintJSON:
				// JSON output uses the same findings slice as the text path so both
				// formats stay in lockstep.
				payload, err := report.JSON(result.Findings)
				if err != nil {
					return NewToolError("rendering JSON output failed", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
			default:
				text := report.Text(result.Findings)
				if text != "" {
					fmt.Fprint(cmd.OutOrStdout(), text)
				}
			}

			if len(result.Findings) > 0 {
				return ExitError{Code: ExitFindings}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&lintFix, "fix", false, "Apply safe formatting fixes before reporting")
	flags.BoolVar(&lintJSON, "json", false, "Render findings as JSON")
	flags.StringArrayVar(&selection.OnlyRules, "only", nil, "Run only the specified rule ID. Can be repeated.")
	flags.StringArrayVar(&selection.SkipRules, "skip", nil, "Skip the specified rule ID. Can be repeated.")

	registerLintFlagCompletions(cmd)

	return cmd
}
