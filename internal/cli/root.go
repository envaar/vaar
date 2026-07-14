/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package cli wires Cobra commands, output rendering and exit codes for
// vaar.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var buildVersion = "dev"

// newRootCmd assembles the root command, keeps the version string in one
// place and turns off Cobra's default usage and error spam so callers see one
// controlled failure path.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vaar",
		Short:   "Repo-aware linter for environment variables",
		Version: buildVersion,
		Long: `Vaar is a repo-aware linter for environment variables.

The current release focuses on deterministic dotenv hygiene, safe fixes,
and automation-friendly output.`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetVersionTemplate("vaar version {{.Version}}\n")

	cmd.AddCommand(newLintCmd())
	return cmd
}

// Execute runs the root command, prints only unexpected failures to stderr,
// and turns Cobra's errors into process exit codes.
func Execute() int {
	cmd := newRootCmd()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	if err := cmd.Execute(); err != nil {
		if !IsExitError(err) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return ExitCode(err)
	}

	return ExitOK
}
