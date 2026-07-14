/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

// Finding records one rule violation and the data needed to render it.
type Finding struct {
	// Rule identifies the rule that raised the finding.
	Rule string `json:"rule"`
	// Severity tells automation how strongly to treat the finding.
	Severity Severity `json:"severity"`
	// File names the file that triggered the finding.
	File string `json:"file"`
	// Line records the 1-based line number for the finding.
	Line int `json:"line"`
	// Message explains the issue in user-facing language.
	Message string `json:"message"`
	// Fixed reports that --fix repaired this finding and it is no longer present.
	Fixed bool `json:"fixed,omitempty"`
}
