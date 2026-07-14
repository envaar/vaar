/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package cli tests completion script generation and rule-ID completion wiring
// for the public command surface.
package cli

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/lint/rules"
	"github.com/spf13/cobra"
)

func TestCompletionCommandGeneratesShellScripts(t *testing.T) {
	cases := map[string][]string{
		"bash": {
			"# bash completion V2 for vaar",
			"__vaar_get_completion_results",
			"__start_vaar()",
		},
		"zsh": {
			"#compdef vaar",
			"compdef _vaar vaar",
			"_vaar()",
		},
		"fish": {
			"# fish completion for vaar",
			"function __vaar_perform_completion",
			"complete -c vaar -e",
		},
		"powershell": {
			"# powershell completion for vaar",
			"[scriptblock]${__vaarCompleterBlock}",
			"Register-ArgumentCompleter -CommandName 'vaar'",
		},
	}

	for shell, snippets := range cases {
		t.Run(shell, func(t *testing.T) {
			var stdout bytes.Buffer

			cmd := newRootCmd()
			cmd.SetOut(&stdout)
			cmd.SetErr(io.Discard)
			cmd.SetArgs([]string{"completion", shell})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("completion command failed: %v", err)
			}

			output := stdout.String()
			if output == "" {
				t.Fatal("expected completion script output")
			}
			for _, snippet := range snippets {
				if !strings.Contains(output, snippet) {
					t.Fatalf("completion script for %s missing %q", shell, snippet)
				}
			}
		})
	}
}

func TestLintFlagCompletionsReturnRuleIDs(t *testing.T) {
	lintCmd, _, err := newRootCmd().Find([]string{"lint"})
	if err != nil {
		t.Fatalf("find lint command failed: %v", err)
	}

	want := ruleIDs()
	for _, flagName := range []string{"only", "skip"} {
		completionFunc, ok := lintCmd.GetFlagCompletionFunc(flagName)
		if !ok {
			t.Fatalf("missing completion function for --%s", flagName)
		}

		completions, directive := completionFunc(lintCmd, nil, "")
		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Fatalf("unexpected completion directive for --%s: got %v want %v", flagName, directive, cobra.ShellCompDirectiveNoFileComp)
		}

		got := make([]string, 0, len(completions))
		for _, completion := range completions {
			got = append(got, string(completion))
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected completions for --%s: got %v want %v", flagName, got, want)
		}
	}
}

func TestLintCompletionRequestReturnsRuleIDs(t *testing.T) {
	want := ruleIDs()

	cases := []struct {
		name string
		args []string
	}{
		{name: "only", args: []string{"__complete", "lint", "--only", ""}},
		{name: "skip", args: []string{"__complete", "lint", "--skip", ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer

			cmd := newRootCmd()
			cmd.SetOut(&stdout)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tc.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("completion request failed: %v", err)
			}

			lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
			if got, want := len(lines), len(want)+1; got != want {
				t.Fatalf("unexpected completion line count: got %d want %d", got, want)
			}

			if got := lines[len(lines)-1]; got != ":4" {
				t.Fatalf("unexpected completion directive: got %q want %q", got, ":4")
			}

			got := lines[:len(lines)-1]
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("unexpected completion results: got %v want %v", got, want)
			}
		})
	}
}

func ruleIDs() []string {
	allRules := rules.All()
	ids := make([]string, 0, len(allRules))
	for _, rule := range allRules {
		ids = append(ids, rule.ID())
	}
	return ids
}
