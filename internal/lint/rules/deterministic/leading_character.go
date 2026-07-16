/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import "github.com/envaar/vaar/internal/lint"

type leadingCharacterRule struct{}

// NewLeadingCharacter returns a rule that warns when a line starts with
// indentation, which often hides accidental formatting drift.
func NewLeadingCharacter() lint.Rule { return leadingCharacterRule{} }

func (leadingCharacterRule) ID() string          { return "leading-character" }
func (leadingCharacterRule) Description() string { return "warns when a line starts with indentation" }

func (leadingCharacterRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if line.IsBlank || line.IsComment || line.LeadingWhitespace == "" {
				continue
			}
			findings = append(findings, finding(
				leadingCharacterRule{}.ID(),
				lint.SeverityWarn,
				file.Path,
				line.Number,
				"line starts with leading whitespace",
			))
		}
	}
	return findings, nil
}
