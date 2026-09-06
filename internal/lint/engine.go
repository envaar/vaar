/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/envaar/vaar/internal/analysis"
)

// EngineOptions contains only rule-selection input for one analysis-backed
// lint execution. Scope, filesystem, output and mutation options belong to
// application orchestration instead.
type EngineOptions struct {
	// OnlyRules keeps only the named rule IDs.
	OnlyRules []string
	// SkipRules removes the named rule IDs after selection.
	SkipRules []string
}

// Engine executes lint rules against an immutable analysis snapshot.
type Engine struct {
	rules []Rule
}

// NewEngine copies the provided rules so callers can reuse or mutate their
// input slice without affecting future runs.
func NewEngine(rules ...Rule) *Engine {
	copied := make([]Rule, len(rules))
	copy(copied, rules)
	return &Engine{rules: copied}
}

// Run selects and executes rules against snapshot. It performs no filesystem
// I/O, parsing, mutation, rendering or exit-code mapping.
func (e *Engine) Run(ctx context.Context, snapshot analysis.Snapshot, opts EngineOptions) ([]Finding, error) {
	selected, err := selectRules(e.rules, opts.OnlyRules, opts.SkipRules)
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0, 16)
	ruleContext := Context{Snapshot: snapshot}
	for _, rule := range selected {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		ruleFindings, err := rule.Run(ruleContext)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rule.ID(), err)
		}
		findings = append(findings, ruleFindings...)
	}

	sortFindings(findings)
	return findings, nil
}

// selectRules applies --only and --skip in declaration order so the selected
// set stays predictable for completions, tests and report output.
func selectRules(all []Rule, only, skip []string) ([]Rule, error) {
	allowed := make(map[string]Rule, len(all))
	ordered := make([]Rule, 0, len(all))
	for _, rule := range all {
		allowed[rule.ID()] = rule
		ordered = append(ordered, rule)
	}

	for _, id := range only {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("invalid empty rule ID in --only")
		}
		if _, ok := allowed[id]; !ok {
			return nil, fmt.Errorf("unknown lint rule %q", id)
		}
	}
	for _, id := range skip {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("invalid empty rule ID in --skip")
		}
		if _, ok := allowed[id]; !ok {
			return nil, fmt.Errorf("unknown lint rule %q", id)
		}
	}

	if len(all) == 0 {
		return nil, nil
	}

	selectedIDs := make(map[string]struct{}, len(all))

	if len(only) > 0 {
		for _, id := range only {
			selectedIDs[strings.TrimSpace(id)] = struct{}{}
		}
	} else {
		for _, rule := range ordered {
			selectedIDs[rule.ID()] = struct{}{}
		}
	}

	if len(skip) > 0 {
		for _, id := range skip {
			delete(selectedIDs, strings.TrimSpace(id))
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
