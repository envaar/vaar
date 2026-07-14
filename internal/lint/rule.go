/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package lint discovers dotenv files, runs rule sets and returns findings in
// a stable order for humans and automation.
package lint

import "github.com/envaar/vaar/internal/envfile"

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

// Context gives each rule read-only access to the repository snapshot and the
// selected command-line options.
type Context struct {
	// Root is the absolute repository root for the current run.
	Root string
	// Files is the parsed file set the runner loaded for this run.
	Files []envfile.File
	// Options holds the resolved lint settings.
	Options Options
}

// Rule describes one lint check that can evaluate the current Context without
// mutating shared state.
type Rule interface {
	ID() string
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
