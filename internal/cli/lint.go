/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"
	"os"
	"path/filepath"

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
				return NewToolError("validating flags", fmt.Errorf("--output requires --json"))
			}

			runner := lint.NewRunner(rules.All()...)
			result, err := runner.Run(cmd.Context(), lint.Options{
				Root:      ".",
				Target:    selection.Target,
				TargetDir: selection.TargetDir,
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
				if lintOutput != "" {
					if err := writeJSONOutputFile(lintOutput, payload); err != nil {
						return NewToolError("writing JSON output file failed", err)
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

			if len(result.Findings) > 0 {
				return ExitError{Code: ExitFindings}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&lintFix, "fix", false, "Apply safe formatting fixes before reporting")
	flags.BoolVar(&lintJSON, "json", false, "Render findings as JSON")
	flags.StringVar(&lintOutput, "output", "", "Write JSON findings to a file (requires --json)")
	flags.StringVar(&selection.Target, "target", "", "Lint only the specified file path.")
	flags.StringVar(&selection.TargetDir, "target-dir", "", "Recursively lint dotenv files under the specified directory.")
	flags.StringArrayVar(&selection.OnlyRules, "only", nil, "Run only the specified rule ID. Can be repeated.")
	flags.StringArrayVar(&selection.SkipRules, "skip", nil, "Skip the specified rule ID. Can be repeated.")

	registerLintFlagCompletions(cmd)

	return cmd
}

func writeJSONOutputFile(path string, payload []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}

	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(append(payload, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
