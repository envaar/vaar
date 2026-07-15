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
	"github.com/envaar/vaar/internal/fs"
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

// Run resolves the root, discovers dotenv files and executes the selected
// rules in declaration order. When fixes are requested, it reports the
// original findings that disappeared as fixed and keeps the post-fix findings.
func (r *Runner) Run(ctx context.Context, opts Options) (Result, error) {
	selected, err := selectRules(r.rules, opts.OnlyRules, opts.SkipRules)
	if err != nil {
		return Result{}, err
	}

	if opts.Root == "" {
		opts.Root = "."
	}

	absRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		return Result{}, fmt.Errorf("resolve root %q: %w", opts.Root, err)
	}

	paths, err := discoverPaths(absRoot, opts.Root, opts.Target, opts.TargetDir)
	if err != nil {
		return Result{}, err
	}

	files, err := loadFiles(absRoot, paths)
	if err != nil {
		return Result{}, err
	}

	findings, err := r.runRules(ctx, absRoot, selected, files, opts)
	if err != nil {
		return Result{}, err
	}

	changed := false
	if opts.Fix {
		changed, err = ApplyFixes(paths)
		if err != nil {
			return Result{}, err
		}

		fixedFiles, err := loadFiles(absRoot, paths)
		if err != nil {
			return Result{}, err
		}

		remaining, err := r.runRules(ctx, absRoot, selected, fixedFiles, opts)
		if err != nil {
			return Result{}, err
		}

		findings = markFixedFindings(findings, remaining)
		files = fixedFiles
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

// discoverPaths resolves the active lint scope and keeps the default recursive
// repository walk unchanged when no explicit target flags are provided.
func discoverPaths(root, rootLabel, target, targetDir string) ([]string, error) {
	if target != "" && targetDir != "" {
		return nil, fmt.Errorf("--target and --target-dir cannot be used together")
	}

	if target != "" {
		path := resolvePath(root, target)
		info, err := os.Stat(path)
		if err != nil {
			return nil, scopePathError("--target", target, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("--target must point to a file: %s", target)
		}
		return []string{path}, nil
	}

	if targetDir != "" {
		path := resolvePath(root, targetDir)
		info, err := os.Stat(path)
		if err != nil {
			return nil, scopePathError("--target-dir", targetDir, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("--target-dir must point to a directory: %s", targetDir)
		}

		paths, err := fs.Discover(path)
		if err != nil {
			return nil, fmt.Errorf("discovering files under %q: %w", targetDir, err)
		}
		return paths, nil
	}

	paths, err := fs.Discover(root)
	if err != nil {
		return nil, fmt.Errorf("discovering files under %q: %w", rootLabel, err)
	}
	return paths, nil
}

// resolvePath turns a user-supplied relative or absolute scope path into an
// absolute path anchored to the current lint root.
func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}

// scopePathError keeps target scope failures easy to understand while
// preserving the underlying OS error for troubleshooting.
func scopePathError(flag, path string, err error) error {
	if os.IsNotExist(err) {
		return fmt.Errorf("%s path does not exist: %s", flag, path)
	}
	return fmt.Errorf("%s path cannot be read: %s: %w", flag, path, err)
}

// loadFiles reads each discovered path and parses it relative to the configured
// repository root.
func loadFiles(root string, paths []string) ([]envfile.File, error) {
	files := make([]envfile.File, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", displayPath(root, path), err)
		}

		parsed, err := envfile.Parse(displayPath(root, path), data)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", displayPath(root, path), err)
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

// displayPath keeps user-facing paths relative to the configured root when
// possible.
func displayPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
