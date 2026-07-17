/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
	"github.com/envaar/vaar/internal/report"
	"github.com/spf13/cobra"
)

// newLintCmd builds the lint subcommand, wires the built-in rule set and
// exposes the flag surface for safe fixes, JSON output, JSON file export,
// rule selection and explicit discovery scopes.
func newLintCmd() *cobra.Command {
	var selection lint.Options
	var lintFix bool
	var lintJSON bool
	var lintOutput string
	var lintListRules bool

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Run environment lint checks",
		Long: `Run Vaar's lint rules against discovered dotenv files in the current repository.

The command supports safe fixes, JSON output, JSON file export through
--output or -o, repeatable rule selection flags and explicit target scopes.
Use --only to narrow the selected rules, --skip to remove rules after
selection, --target to lint one file and --target-dir to discover files under
one directory. --output writes JSON to a file instead of stdout and requires
--json. Use --list-rules to print every registered rule with its description
without running anything:

  vaar lint --only=duplicate-key
  vaar lint --only=duplicate-key --only=invalid-key-name
  vaar lint --json --output=lint-report.json
  vaar lint --skip=trailing-whitespace
  vaar lint --skip=trailing-whitespace --skip=extra-blank-line
  vaar lint --target=.env.staging
  vaar lint --target-dir=src
  vaar lint --list-rules

Use either --target or --target-dir, not both.`,
		Example: `  vaar lint
  vaar lint --json
  vaar lint --json --output=lint-report.json
  vaar lint --json -o lint-report.json
  vaar lint --fix
  vaar lint --only=duplicate-key --only=invalid-key-name
  vaar lint --skip=trailing-whitespace --skip=extra-blank-line
  vaar lint --target=.env.staging
  vaar lint --target-dir=src
  vaar lint --list-rules`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if lintListRules {
				if conflict := firstListRulesConflict(selection, lintFix, lintJSON, lintOutput); conflict != "" {
					return NewToolError(fmt.Sprintf("--list-rules cannot be combined with %s", conflict), nil)
				}
				return writeRuleList(cmd.OutOrStdout(), rules.All())
			}

			if lintOutput != "" && !lintJSON {
				return NewToolError("--output requires --json", nil)
			}

			opts := lint.Options{
				Root:      ".",
				Target:    selection.Target,
				TargetDir: selection.TargetDir,
				OnlyRules: selection.OnlyRules,
				SkipRules: selection.SkipRules,
				Fix:       lintFix,
			}

			if lintOutput != "" {
				if err := validateOutputDestination(lintOutput); err != nil {
					return err
				}
				if err := lint.ValidateOutputPath(opts, lintOutput); err != nil {
					return NewToolError(err.Error(), nil)
				}
			}

			runner := lint.NewRunner(rules.All()...)
			result, err := runner.Run(cmd.Context(), opts)
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
				if lintOutput != "" {
					if err := writeJSONOutput(lintOutput, append(payload, '\n')); err != nil {
						return err
					}

					count := len(result.Findings)
					if count == 0 {
						fmt.Fprintf(cmd.ErrOrStderr(), "Successfully flagged no findings and wrote to %s\n", lintOutput)
					} else if count == 1 {
						fmt.Fprintf(cmd.ErrOrStderr(), "Successfully flagged 1 finding and wrote to %s\n", lintOutput)
					} else {
						fmt.Fprintf(cmd.ErrOrStderr(), "Successfully flagged %d findings and wrote to %s\n", count, lintOutput)
					}
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				}
			default:
				text := report.Text(result.Findings)
				if text != "" {
					fmt.Fprint(cmd.OutOrStdout(), text)
				}
			}

			if result.HasUnfixedFindings() {
				return ExitError{Code: ExitFindings}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&lintFix, "fix", false, "Apply safe formatting fixes before reporting")
	flags.BoolVar(&lintJSON, "json", false, "Render findings as JSON")
	flags.StringVarP(&lintOutput, "output", "o", "", "Write JSON output to a file instead of stdout (requires --json)")
	flags.StringVar(&selection.Target, "target", "", "Lint only the specified file path.")
	flags.StringVar(&selection.TargetDir, "target-dir", "", "Recursively lint dotenv files under the specified directory.")
	flags.StringArrayVar(&selection.OnlyRules, "only", nil, "Run only the specified rule ID. Can be repeated.")
	flags.StringArrayVar(&selection.SkipRules, "skip", nil, "Skip the specified rule ID. Can be repeated.")
	flags.BoolVar(&lintListRules, "list-rules", false, "List every registered rule and its description, then exit")

	registerLintFlagCompletions(cmd)

	return cmd
}

// firstListRulesConflict returns the first execution flag that conflicts with
// --list-rules, or an empty string when there is no conflict.
func firstListRulesConflict(selection lint.Options, fix, json bool, output string) string {
	if len(selection.OnlyRules) > 0 {
		return "--only"
	}
	if len(selection.SkipRules) > 0 {
		return "--skip"
	}
	if fix {
		return "--fix"
	}
	if output != "" {
		return "--output"
	}
	if json {
		return "--json"
	}
	if selection.Target != "" {
		return "--target"
	}
	if selection.TargetDir != "" {
		return "--target-dir"
	}
	return ""
}

// writeRuleList prints the registered rules in alphabetical order with aligned
// NAME and DESCRIPTION columns. Fixability metadata is not part of the Rule
// interface, so no FIXABLE column is emitted.
func writeRuleList(w io.Writer, all []lint.Rule) error {
	sorted := make([]lint.Rule, len(all))
	copy(sorted, all)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID() < sorted[j].ID()
	})

	nameWidth := len("NAME")
	for _, rule := range sorted {
		if n := len(rule.ID()); n > nameWidth {
			nameWidth = n
		}
	}

	if _, err := fmt.Fprintf(w, "%-*s  DESCRIPTION\n", nameWidth, "NAME"); err != nil {
		return err
	}
	for _, rule := range sorted {
		if _, err := fmt.Fprintf(w, "%-*s  %s\n", nameWidth, rule.ID(), rule.Description()); err != nil {
			return err
		}
	}
	return nil
}
