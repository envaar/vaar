/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/scope"
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
	selection, err := scope.Resolve(scope.Options{
		Root:      opts.Root,
		Target:    opts.Target,
		TargetDir: opts.TargetDir,
	})
	if err != nil {
		return Result{}, err
	}

	return r.runWithSelection(ctx, opts, selection)
}

// RunWithSelection executes the configured rules against a pre-resolved scope
// selection. Callers that already resolved scope, such as the CLI preflight for
// output-path validation, can reuse the same selection here to avoid walking the
// tree twice.
func (r *Runner) RunWithSelection(ctx context.Context, opts Options, selection scope.Selection) (Result, error) {
	return r.runWithSelection(ctx, opts, selection)
}

func (r *Runner) runWithSelection(ctx context.Context, opts Options, selection scope.Selection) (Result, error) {
	selected, err := selectRules(r.rules, opts.OnlyRules, opts.SkipRules)
	if err != nil {
		return Result{}, err
	}

	files, err := loadFiles(selection)
	if err != nil {
		return Result{}, err
	}

	findings, err := r.runRules(ctx, selection.Root, selected, files, opts)
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
			fixedFiles, err := loadFiles(selection)
			if err != nil {
				return Result{}, err
			}

			remaining, err := r.runRules(ctx, selection.Root, selected, fixedFiles, opts)
			if err != nil {
				return Result{}, err
			}

			findings = markFixedFindings(findings, remaining)
			files = fixedFiles
		}
	}

	sortFindings(findings)

	return Result{
		Findings: findings,
		Files:    files,
		Changed:  changed,
	}, nil
}

func (r *Runner) runRules(ctx context.Context, root string, selected []Rule, files []envfile.File, opts Options) ([]Finding, error) {
	runCtx := Context{Root: root, Files: files, Options: opts}
	findings := make([]Finding, 0, 16)
	for _, rule := range selected {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		ruleFindings, err := rule.Run(runCtx)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rule.ID(), err)
		}
		findings = append(findings, ruleFindings...)
	}

	return findings, nil
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

// ValidateOutputPath reports an error if outputPath resolves to the same file
// as any input path in the resolved selection. The comparison uses canonical
// absolute paths so relative, cleaned and symlink-equivalent forms are
// detected.
func ValidateOutputPath(selection scope.Selection, outputPath string) error {
	out, err := canonicalPath(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path %q: %w", outputPath, err)
	}

	for _, input := range selection.Paths {
		in, err := canonicalPath(input)
		if err != nil {
			return fmt.Errorf("resolve input path %q: %w", input, err)
		}
		if out == in {
			return fmt.Errorf("cannot write lint output to %q: the path is also a lint input file", outputPath)
		}
	}

	return nil
}

// canonicalPath returns a stable absolute form of path. It resolves symlinks
// and cleans . and .. segments. If path does not exist yet (typical for an
// --output destination), the parent directory is canonicalized and the base
// name is rejoined so the result can still be compared with existing files.
func canonicalPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved, nil
	}

	dir := filepath.Dir(abs)
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedDir, filepath.Base(abs)), nil
}

// loadFiles reads each selected path and parses it relative to the configured
// repository root.
func loadFiles(selection scope.Selection) ([]envfile.File, error) {
	files := make([]envfile.File, 0, len(selection.Paths))
	for _, path := range selection.Paths {
		display := selection.DisplayPath(path)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", display, err)
		}

		parsed, err := envfile.Parse(display, data)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", display, err)
		}
		files = append(files, parsed)
	}
	return files, nil
}

// selectRules applies --only and --skip in declaration order so the selected
// set stays predictable for completions, tests and report output.
func selectRules(all []Rule, only, skip []string) ([]Rule, error) {
	if len(all) == 0 {
		return nil, nil
	}

	allowed := make(map[string]Rule, len(all))
	ordered := make([]Rule, 0, len(all))
	for _, rule := range all {
		allowed[rule.ID()] = rule
		ordered = append(ordered, rule)
	}

	selectedIDs := make(map[string]struct{}, len(all))

	if len(only) > 0 {
		for _, id := range only {
			id = strings.TrimSpace(id)
			if id == "" {
				return nil, fmt.Errorf("invalid empty rule ID in --only")
			}
			if _, ok := allowed[id]; !ok {
				return nil, fmt.Errorf("unknown lint rule %q", id)
			}
			selectedIDs[id] = struct{}{}
		}
	} else {
		for _, rule := range ordered {
			selectedIDs[rule.ID()] = struct{}{}
		}
	}

	if len(skip) > 0 {
		for _, id := range skip {
			id = strings.TrimSpace(id)
			if id == "" {
				return nil, fmt.Errorf("invalid empty rule ID in --skip")
			}
			if _, ok := allowed[id]; !ok {
				return nil, fmt.Errorf("unknown lint rule %q", id)
			}
			delete(selectedIDs, id)
		}
	}

	selected := make([]Rule, 0, len(selectedIDs))
	for _, rule := range ordered {
		if _, ok := selectedIDs[rule.ID()]; ok {
			selected = append(selected, rule)
		}
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("no lint rules selected after applying --only and --skip")
	}

	return selected, nil
}

// sortFindings orders output by file, line, severity, rule and message so
// repeated runs produce the same report.
func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		left := findings[i]
		right := findings[j]
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Severity.Rank() != right.Severity.Rank() {
			return left.Severity.Rank() < right.Severity.Rank()
		}
		if left.Rule != right.Rule {
			return left.Rule < right.Rule
		}
		return left.Message < right.Message
	})
}
