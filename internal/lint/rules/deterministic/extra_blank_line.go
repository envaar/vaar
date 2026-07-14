/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import "github.com/envaar/vaar/internal/lint"

type extraBlankLineRule struct{}

// NewExtraBlankLine returns a rule that flags repeated blank lines inside a
// file because they add vertical noise without changing meaning.
func NewExtraBlankLine() lint.Rule { return extraBlankLineRule{} }

func (extraBlankLineRule) ID() string { return "extra-blank-line" }

func (extraBlankLineRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		run := 0
		for _, line := range file.Lines {
			if line.IsBlank {
				run++
				if run > 1 {
					findings = append(findings, finding(
						extraBlankLineRule{}.ID(),
						lint.SeverityWarn,
						file.Path,
						line.Number,
						"repeated blank line",
					))
				}
				continue
			}
			run = 0
		}
	}
	return findings, nil
}
