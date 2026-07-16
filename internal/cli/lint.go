/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"
	"os"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
	"github.com/envaar/vaar/internal/report"
	"github.com/spf13/cobra"
)

// newLintCmd builds the lint subcommand, wires the built-in rule set and
// exposes the flag surface for safe fixes, JSON output, rule selection and
// explicit discovery scopes.
func newLintCmd() *cobra.Command {
	var selection lint.Options
	var lintFix bool
	var lintJSON bool
	var lintOutput string

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Run environment lint checks",
		Long: `Run Vaar's lint rules against discovered dotenv files in the current repository.

The command supports safe fixes, JSON output, repeatable rule selection flags
and explicit target scopes. Use --only to narrow the selected rules, --skip to
remove rules after selection, --target to lint one file and --target-dir to
discover files under one directory:

  vaar lint --only=duplicate-key
  vaar lint --only=duplicate-key --only=invalid-key-name
  vaar lint --skip=trailing-whitespace
  vaar lint --skip=trailing-whitespace --skip=extra-blank-line
  vaar lint --target=.env.staging
  vaar lint --target-dir=src

Use either --target or --target-dir, not both.`,
		Example: `  vaar lint
  vaar lint --json
  vaar lint --fix
  vaar lint --only=duplicate-key --only=invalid-key-name
  vaar lint --skip=trailing-whitespace --skip=extra-blank-line
  vaar lint --target=.env.staging
  vaar lint --target-dir=src`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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
					data := append(payload, '\n')
					dir := "."
					for i := len(lintOutput) - 1; i >= 0; i-- {
						if lintOutput[i] == '/' || lintOutput[i] == '\\' {
							if i == 0 {
								dir = lintOutput[:1]
							} else {
								dir = lintOutput[:i]
							}
							break
						}
					}

					tmp, err := os.CreateTemp(dir, "vaar-lint-*.json")
					if err != nil {
						return NewToolError("creating JSON output file failed", err)
					}
					tmpName := tmp.Name()
					defer os.Remove(tmpName)

					if _, err := tmp.Write(data); err != nil {
						_ = tmp.Close()
						return NewToolError(fmt.Sprintf("writing JSON output to %s failed", lintOutput), err)
					}
					if err := tmp.Close(); err != nil {
						return NewToolError(fmt.Sprintf("writing JSON output to %s failed", lintOutput), err)
					}

					if err := os.Rename(tmpName, lintOutput); err != nil {
						// Windows rename does not replace existing files.
						if removeErr := os.Remove(lintOutput); removeErr == nil {
							err = os.Rename(tmpName, lintOutput)
						}
						if err != nil {
							return NewToolError(fmt.Sprintf("writing JSON output to %s failed", lintOutput), err)
						}
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

	registerLintFlagCompletions(cmd)

	return cmd
}
