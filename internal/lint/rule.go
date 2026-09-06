/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package lint runs analysis-backed dotenv rules and returns findings in a
// stable order for humans and automation. The legacy Runner remains a
// temporary orchestration adapter while callers migrate to the layered path.
package lint

import (
	"github.com/envaar/vaar/internal/analysis"
	"github.com/envaar/vaar/internal/envfile"
)

// Options control one lint run without tying callers to Cobra flags. The CLI
// maps its flags into this struct before handing work to the runner.
type Options struct {
	// Root is the repository root to scan.
	Root string
	// Target limits a run to one explicit file path.
	Target string
	// TargetDir limits a run to dotenv files discovered under one directory.
	TargetDir string
	// OnlyRules keeps only the named rule IDs.
	OnlyRules []string
	// SkipRules removes the named rule IDs after selection.
	SkipRules []string
	// Fix enables the safe formatting pass and reports findings repaired by it.
	Fix bool
}

// Context gives each rule read-only access to the immutable analysis snapshot.
// Scope, filesystem, command-line and mutation concerns are intentionally not
// part of the rule-facing context.
type Context struct {
	// Snapshot contains the ordered, value-free documents being linted.
	Snapshot analysis.Snapshot
}

// Rule describes one lint check that can evaluate the current Context without
// mutating shared state. FixableRule remains a separate byte-transform
// contract for the temporary compatibility runner.
type Rule interface {
	ID() string
	Description() string
	Run(Context) ([]Finding, error)
}

// Result groups the final findings, the parsed files and the fix status.
type Result struct {
	// Findings contains the sorted lint output.
	Findings []Finding
	// Files contains the parsed file set used to generate the findings.
	Files []envfile.File
	// Changed reports whether ApplyFixes rewrote any file on disk.
	Changed bool
}

// HasUnfixedFindings reports whether the final lint snapshot still contains
// any findings after an optional fix pass.
func (r Result) HasUnfixedFindings() bool {
	for _, finding := range r.Findings {
		if !finding.Fixed {
			return true
		}
	}
	return false
}
