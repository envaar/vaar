/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"github.com/envaar/vaar/internal/analysis"
	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
)

type extraBlankLineRule struct{}

// NewExtraBlankLine returns a rule that flags repeated blank lines inside a
// file because they add vertical noise without changing meaning.
func NewExtraBlankLine() lint.Rule { return extraBlankLineRule{} }

func (extraBlankLineRule) ID() string          { return "extra-blank-line" }
func (extraBlankLineRule) Description() string { return "flags repeated blank lines inside a file" }

// Fix collapses each run of consecutive blank lines to a single blank line.
func (extraBlankLineRule) Fix(data []byte) []byte { return envfile.CollapseBlankLines(data) }

func (extraBlankLineRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	ctx.Snapshot.RangeDocuments(func(document analysis.DocumentView) {
		run := 0
		document.RangeLines(func(line analysis.Line) {
			if line.IsBlank {
				run++
				if run > 1 {
					findings = append(findings, finding(
						extraBlankLineRule{}.ID(),
						lint.SeverityWarn,
						document.DisplayPath,
						line.Number,
						"repeated blank line",
					))
				}
				return
			}
			run = 0
		})
	})
	return findings, nil
}
