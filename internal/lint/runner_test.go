/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package lint_test verifies rule selection, execution order and discovery
// behavior in the runner.
package lint_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/cli"
	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
)

type countingRule struct {
	id    string
	calls *int
}

func (r countingRule) ID() string { return r.id }

func (r countingRule) Run(lint.Context) ([]lint.Finding, error) {
	*r.calls++
	return nil, nil
}

func TestRunnerRuleSelection(t *testing.T) {
	root := t.TempDir()

	ruleset, counts := testRules()
	runner := lint.NewRunner(ruleset...)

	cases := []struct {
		name        string
		only        []string
		skip        []string
		wantCounts  map[string]int
		wantErrText string
	}{
		{
			name:       "no only and no skip runs all rules",
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 1, "quote-character": 1, "trailing-whitespace": 1, "extra-blank-line": 1},
		},
		{
			name:       "only runs exact rule",
			only:       []string{"duplicate-key"},
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 0, "quote-character": 0, "trailing-whitespace": 0, "extra-blank-line": 0},
		},
		{
			name:       "repeated only runs the union",
			only:       []string{"duplicate-key", "invalid-key-name", "quote-character"},
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 1, "quote-character": 1, "trailing-whitespace": 0, "extra-blank-line": 0},
		},
		{
			name:       "duplicate only is de-duped",
			only:       []string{"duplicate-key", "duplicate-key"},
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 0, "quote-character": 0, "trailing-whitespace": 0, "extra-blank-line": 0},
		},
		{
			name:       "skip removes a rule",
			skip:       []string{"trailing-whitespace"},
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 1, "quote-character": 1, "trailing-whitespace": 0, "extra-blank-line": 1},
		},
		{
			name:       "repeated skip removes multiple rules",
			skip:       []string{"trailing-whitespace", "extra-blank-line"},
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 1, "quote-character": 1, "trailing-whitespace": 0, "extra-blank-line": 0},
		},
		{
			name:       "duplicate skip is de-duped",
			skip:       []string{"trailing-whitespace", "trailing-whitespace"},
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 1, "quote-character": 1, "trailing-whitespace": 0, "extra-blank-line": 1},
		},
		{
			name:       "only and skip interact in order",
			only:       []string{"duplicate-key", "quote-character"},
			skip:       []string{"quote-character"},
			wantCounts: map[string]int{"duplicate-key": 1, "invalid-key-name": 0, "quote-character": 0, "trailing-whitespace": 0, "extra-blank-line": 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetCounts(counts)

			result, err := runner.Run(context.Background(), lint.Options{
				Root:      root,
				OnlyRules: tc.only,
				SkipRules: tc.skip,
			})
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			if len(result.Findings) != 0 {
				t.Fatalf("expected no findings, got %d", len(result.Findings))
			}

			assertCounts(t, counts, tc.wantCounts)
		})
	}
}

func TestRunnerRuleSelectionErrors(t *testing.T) {
	root := t.TempDir()
	ruleset, counts := testRules()
	runner := lint.NewRunner(ruleset...)

	cases := []struct {
		name        string
		only        []string
		skip        []string
		wantErrText string
	}{
		{
			name:        "final empty rule set returns a tool error",
			only:        []string{"duplicate-key"},
			skip:        []string{"duplicate-key"},
			wantErrText: "no lint rules selected after applying --only and --skip",
		},
		{
			name:        "unknown only rule returns exit code 2",
			only:        []string{"unknown-rule"},
			wantErrText: "unknown lint rule \"unknown-rule\"",
		},
		{
			name:        "unknown skip rule returns exit code 2",
			skip:        []string{"unknown-rule"},
			wantErrText: "unknown lint rule \"unknown-rule\"",
		},
		{
			name:        "json output is not a rule id",
			only:        []string{"json-output"},
			wantErrText: "unknown lint rule \"json-output\"",
		},
		{
			name:        "masked output is not a rule id",
			skip:        []string{"masked-output"},
			wantErrText: "unknown lint rule \"masked-output\"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetCounts(counts)

			result, err := runner.Run(context.Background(), lint.Options{
				Root:      root,
				OnlyRules: tc.only,
				SkipRules: tc.skip,
			})
			if err == nil {
				t.Fatalf("expected error, got result: %#v", result)
			}

			if got := cli.ExitCode(err); got != cli.ExitInternal {
				t.Fatalf("unexpected exit code: got %d want %d", got, cli.ExitInternal)
			}
			if !strings.Contains(err.Error(), tc.wantErrText) {
				t.Fatalf("unexpected error: %v", err)
			}

			assertCounts(t, counts, map[string]int{
				"duplicate-key":       0,
				"invalid-key-name":    0,
				"quote-character":     0,
				"trailing-whitespace": 0,
				"extra-blank-line":    0,
			})
		})
	}
}

func TestRunnerOnlySelectionStillUsesRealRules(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("KEY=value\nKEY=value-2\nTRAIL=one  \n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	runner := lint.NewRunner(rules.All()...)
	result, err := runner.Run(context.Background(), lint.Options{
		Root:      root,
		OnlyRules: []string{"duplicate-key"},
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if got, want := len(result.Findings), 1; got != want {
		t.Fatalf("unexpected findings count: got %d want %d", got, want)
	}
	if got, want := result.Findings[0].Rule, "duplicate-key"; got != want {
		t.Fatalf("unexpected rule: got %q want %q", got, want)
	}
}

func TestRunnerTargetFileModeUsesExplicitPath(t *testing.T) {
	root := t.TempDir()

	mustWrite := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	mustWrite(filepath.Join(root, ".env"), "ROOT=value\nROOT=other\n")
	mustWrite(filepath.Join(root, ".env.staging"), "STAGING=value\nSTAGING=other\n")

	runner := lint.NewRunner(rules.All()...)
	result, err := runner.Run(context.Background(), lint.Options{
		Root:   root,
		Target: ".env.staging",
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if got, want := len(result.Files), 1; got != want {
		t.Fatalf("unexpected file count: got %d want %d", got, want)
	}
	if got, want := result.Files[0].Path, ".env.staging"; got != want {
		t.Fatalf("unexpected parsed path: got %q want %q", got, want)
	}
	if got, want := len(result.Findings), 1; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
	if got, want := result.Findings[0].File, ".env.staging"; got != want {
		t.Fatalf("unexpected finding file: got %q want %q", got, want)
	}
}

func TestRunnerTargetDirModeRecursesAndSkipsIgnoredDirs(t *testing.T) {
	root := t.TempDir()

	mustWrite := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	mustWrite(filepath.Join(root, ".env"), "ROOT=value\nROOT=other\n")
	mustWrite(filepath.Join(root, "src", "app", ".env.example"), "APP=value\nAPP=other\n")
	mustWrite(filepath.Join(root, "src", "examples", "broken", ".env.example"), "EXAMPLE=value\nEXAMPLE=other\n")
	mustWrite(filepath.Join(root, "src", "dist", ".env.local"), "DIST=value\nDIST=other\n")

	runner := lint.NewRunner(rules.All()...)
	result, err := runner.Run(context.Background(), lint.Options{
		Root:      root,
		TargetDir: "src",
	})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	wantPaths := []string{
		filepath.Join("src", "app", ".env.example"),
		filepath.Join("src", "examples", "broken", ".env.example"),
	}

	if got, want := len(result.Files), len(wantPaths); got != want {
		t.Fatalf("unexpected file count: got %d want %d", got, want)
	}
	for i, wantPath := range wantPaths {
		if got, want := result.Files[i].Path, wantPath; got != want {
			t.Fatalf("unexpected parsed path at %d: got %q want %q", i, got, want)
		}
	}
	if got, want := len(result.Findings), len(wantPaths); got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
	for i, wantPath := range wantPaths {
		if got, want := result.Findings[i].File, wantPath; got != want {
			t.Fatalf("unexpected finding file at %d: got %q want %q", i, got, want)
		}
	}
}

func testRules() ([]lint.Rule, map[string]*int) {
	counts := map[string]*int{
		"duplicate-key":       new(int),
		"invalid-key-name":    new(int),
		"quote-character":     new(int),
		"trailing-whitespace": new(int),
		"extra-blank-line":    new(int),
	}

	return []lint.Rule{
		countingRule{id: "duplicate-key", calls: counts["duplicate-key"]},
		countingRule{id: "invalid-key-name", calls: counts["invalid-key-name"]},
		countingRule{id: "quote-character", calls: counts["quote-character"]},
		countingRule{id: "trailing-whitespace", calls: counts["trailing-whitespace"]},
		countingRule{id: "extra-blank-line", calls: counts["extra-blank-line"]},
	}, counts
}

func resetCounts(counts map[string]*int) {
	for _, count := range counts {
		*count = 0
	}
}

func assertCounts(t *testing.T, counts map[string]*int, want map[string]int) {
	t.Helper()

	for id, expected := range want {
		if got := *counts[id]; got != expected {
			t.Fatalf("unexpected call count for %s: got %d want %d", id, got, expected)
		}
	}
}
