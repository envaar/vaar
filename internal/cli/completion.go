/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"

	"github.com/envaar/vaar/internal/lint/rules"
	"github.com/spf13/cobra"
)

// registerLintFlagCompletions keeps `--only` and `--skip` aligned with the
// built-in rule IDs and disables file completion for those flags.
func registerLintFlagCompletions(cmd *cobra.Command) {
	completionFunc := cobra.FixedCompletions(lintRuleCompletions(), cobra.ShellCompDirectiveNoFileComp)

	mustRegisterFlagCompletionFunc(cmd, "only", completionFunc)
	mustRegisterFlagCompletionFunc(cmd, "skip", completionFunc)
}

// lintRuleCompletions returns the current rule IDs in declaration order so
// shell completion stays stable as the rule set grows.
func lintRuleCompletions() []cobra.Completion {
	allRules := rules.All()
	completions := make([]cobra.Completion, 0, len(allRules))
	for _, rule := range allRules {
		completions = append(completions, cobra.Completion(rule.ID()))
	}
	return completions
}

// mustRegisterFlagCompletionFunc fails fast because a broken completion hook
// is a startup bug, not a runtime condition.
func mustRegisterFlagCompletionFunc(cmd *cobra.Command, flagName string, fn cobra.CompletionFunc) {
	if err := cmd.RegisterFlagCompletionFunc(flagName, fn); err != nil {
		panic(fmt.Sprintf("register completion for --%s: %v", flagName, err))
	}
}
