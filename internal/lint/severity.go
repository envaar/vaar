/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

// Severity describes how much a finding should affect automation.
type Severity string

const (
	// SeverityError marks findings that should fail automation.
	SeverityError Severity = "error"
	// SeverityWarn marks findings that should be surfaced without failing.
	SeverityWarn Severity = "warn"
)

// Rank returns a sortable weight that puts errors ahead of warnings.
func (s Severity) Rank() int {
	switch s {
	case SeverityError:
		return 0
	case SeverityWarn:
		return 1
	default:
		return 2
	}
}
