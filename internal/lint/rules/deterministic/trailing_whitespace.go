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

type trailingWhitespaceRule struct{}

// NewTrailingWhitespace returns a rule that warns about trailing whitespace,
// which tends to sneak into review diffs unnoticed.
func NewTrailingWhitespace() lint.Rule { return trailingWhitespaceRule{} }

func (trailingWhitespaceRule) ID() string          { return "trailing-whitespace" }
func (trailingWhitespaceRule) Description() string { return "warns about trailing whitespace" }

// Fix trims trailing spaces and tabs from every line and empties whitespace-only lines.
func (trailingWhitespaceRule) Fix(data []byte) []byte { return envfile.TrimTrailingWhitespace(data) }

func (trailingWhitespaceRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	ctx.Snapshot.RangeDocuments(func(document analysis.DocumentView) {
		document.RangeLines(func(line analysis.Line) {
			if !line.HasTrailingWhitespace {
				return
			}
			findings = append(findings, finding(
				trailingWhitespaceRule{}.ID(),
				lint.SeverityWarn,
				document.DisplayPath,
				line.Number,
				"line has trailing whitespace",
			))
		})
	})
	return findings, nil
}
