/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

import (
	"context"
	"fmt"

	"github.com/envaar/vaar/internal/analysis"
	analysisdotenv "github.com/envaar/vaar/internal/analysis/dotenv"
	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/fs"
	"github.com/envaar/vaar/internal/scope"
	sourcedotenv "github.com/envaar/vaar/internal/source/dotenv"
)

// Runner executes a rule set over discovered dotenv files and keeps the
// resulting report deterministic.
type Runner struct {
	rules []Rule
}

// NewRunner copies the provided rules so callers can reuse or mutate their
// input slice without affecting future runs.
func NewRunner(rules ...Rule) *Runner {
	copied := make([]Rule, len(rules))
	copy(copied, rules)
	return &Runner{rules: copied}
}

// ValidateRuleSelection checks the requested only/skip rule IDs against the
// available rule set and returns the same validation errors that Run would.
func ValidateRuleSelection(all []Rule, only, skip []string) error {
	_, err := selectRules(all, only, skip)
	return err
}

// Run resolves the root, discovers dotenv files and executes the selected
// rules in declaration order. When fixes are requested, it reports the
// original findings that disappeared as fixed and keeps the post-fix findings.
func (r *Runner) Run(ctx context.Context, opts Options) (Result, error) {
	opts = normalizeOptions(opts)

	selected, err := selectRules(r.rules, opts.OnlyRules, opts.SkipRules)
	if err != nil {
		return Result{}, err
	}

	selection, err := scope.Resolve(scope.Options{
		Root:      opts.Root,
		Target:    opts.Target,
		TargetDir: opts.TargetDir,
	})
	if err != nil {
		return Result{}, err
	}

	return r.runWithSelection(ctx, opts, selection, selected)
}

// RunWithSelection executes the configured rules against a pre-resolved scope
// selection. Callers that already resolved scope, such as the CLI preflight for
// output-path validation, can reuse the same selection here to avoid walking the
// tree twice.
func (r *Runner) RunWithSelection(ctx context.Context, opts Options, selection scope.Selection) (Result, error) {
	opts = normalizeOptions(opts)

	selected, err := selectRules(r.rules, opts.OnlyRules, opts.SkipRules)
	if err != nil {
		return Result{}, err
	}

	return r.runWithSelection(ctx, opts, selection, selected)
}

func (r *Runner) runWithSelection(ctx context.Context, opts Options, selection scope.Selection, selected []Rule) (Result, error) {
	loaded, err := loadFiles(selection)
	if err != nil {
		return Result{}, err
	}

	findings, err := runRules(ctx, selected, loaded.snapshot)
	if err != nil {
		return Result{}, err
	}

	changed := false
	if opts.Fix {
		changed, err = ApplyFixes(selected, selection.Paths)
		if err != nil {
			return Result{}, err
		}

		if changed {
			fixed, err := loadFiles(selection)
			if err != nil {
				return Result{}, err
			}

			remaining, err := runRules(ctx, selected, fixed.snapshot)
			if err != nil {
				return Result{}, err
			}

			findings = markFixedFindings(findings, remaining)
			loaded = fixed
		}
	}

	sortFindings(findings)

	return Result{
		Findings: findings,
		Files:    loaded.files,
		Changed:  changed,
	}, nil
}

func normalizeOptions(opts Options) Options {
	if opts.Root == "" {
		opts.Root = "."
	}
	return opts
}

func runRules(ctx context.Context, selected []Rule, snapshot analysis.Snapshot) ([]Finding, error) {
	return NewEngine(selected...).Run(ctx, snapshot, EngineOptions{})
}

type findingKey struct {
	rule     string
	severity Severity
	file     string
	message  string
}

func keyForFinding(finding Finding) findingKey {
	return findingKey{
		rule:     finding.Rule,
		severity: finding.Severity,
		file:     finding.File,
		message:  finding.Message,
	}
}

func markFixedFindings(original, remaining []Finding) []Finding {
	remainingCounts := make(map[findingKey]int, len(remaining))
	for _, finding := range remaining {
		remainingCounts[keyForFinding(finding)]++
	}

	findings := make([]Finding, 0, len(original)+len(remaining))
	for _, finding := range original {
		key := keyForFinding(finding)
		if remainingCounts[key] > 0 {
			remainingCounts[key]--
			continue
		}

		finding.Fixed = true
		findings = append(findings, finding)
	}

	return append(findings, remaining...)
}

type loadedFiles struct {
	files    []envfile.File
	snapshot analysis.Snapshot
}

// loadFiles reads each selected path and parses it relative to the configured
// repository root. The parsed files remain available only for the temporary
// runner result and fix adapter; rules receive the derived analysis snapshot.
func loadFiles(selection scope.Selection) (loadedFiles, error) {
	files := make([]envfile.File, 0, len(selection.Paths))
	inputs := make([]analysisdotenv.DocumentInput, 0, len(selection.Paths))
	for _, path := range selection.Paths {
		display := selection.DisplayPath(path)
		data, err := fs.ReadFile(path)
		if err != nil {
			return loadedFiles{}, fmt.Errorf("read %q: %w", display, err)
		}

		parsed, err := envfile.Parse(display, data)
		if err != nil {
			return loadedFiles{}, fmt.Errorf("parse %q: %w", display, err)
		}
		files = append(files, parsed)
		inputs = append(inputs, analysisdotenv.DocumentInput{
			ID: analysis.DocumentID(path),
			Source: sourcedotenv.Document{
				File:       parsed,
				SourcePath: path,
			},
		})
	}

	return loadedFiles{
		files:    files,
		snapshot: analysis.NewSnapshot(analysisdotenv.FromDocuments(inputs)),
	}, nil
}
