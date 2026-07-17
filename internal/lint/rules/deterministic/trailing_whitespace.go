/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import "github.com/envaar/vaar/internal/lint"

type trailingWhitespaceRule struct{}

// NewTrailingWhitespace returns a rule that warns about trailing whitespace,
// which tends to sneak into review diffs unnoticed.
func NewTrailingWhitespace() lint.Rule { return trailingWhitespaceRule{} }

func (trailingWhitespaceRule) ID() string          { return "trailing-whitespace" }
func (trailingWhitespaceRule) Description() string { return "warns about trailing whitespace" }

func (trailingWhitespaceRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if line.TrailingWhitespace == "" {
				continue
			}
			findings = append(findings, finding(
				trailingWhitespaceRule{}.ID(),
				lint.SeverityWarn,
				file.Path,
				line.Number,
				"line has trailing whitespace",
			))
		}
	}
	return findings, nil
}
