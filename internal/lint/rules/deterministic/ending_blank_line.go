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

type endingBlankLineRule struct{}

// NewEndingBlankLine returns a rule that flags files that do not end with one
// clean trailing newline.
func NewEndingBlankLine() lint.Rule { return endingBlankLineRule{} }

func (endingBlankLineRule) ID() string { return "ending-blank-line" }
func (endingBlankLineRule) Description() string {
	return "flags files that do not end with one clean trailing newline"
}

// Fix removes trailing blank lines and leaves exactly one final newline.
func (endingBlankLineRule) Fix(data []byte) []byte { return envfile.TrimFinalBlankLines(data) }

func (endingBlankLineRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	ctx.Snapshot.RangeDocuments(func(document analysis.DocumentView) {
		var last analysis.Line
		hasLine := false
		document.RangeLines(func(line analysis.Line) {
			last = line
			hasLine = true
		})
		if !hasLine {
			return
		}
		if !document.EndsWithNewline || last.IsBlank {
			findings = append(findings, finding(
				endingBlankLineRule{}.ID(),
				lint.SeverityWarn,
				document.DisplayPath,
				last.Number,
				"file must end with exactly one final newline",
			))
		}
	})
	return findings, nil
}
