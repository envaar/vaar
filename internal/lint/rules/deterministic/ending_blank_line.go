/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import "github.com/envaar/vaar/internal/lint"

type endingBlankLineRule struct{}

// NewEndingBlankLine returns a rule that flags files that do not end with one
// clean trailing newline.
func NewEndingBlankLine() lint.Rule { return endingBlankLineRule{} }

func (endingBlankLineRule) ID() string { return "ending-blank-line" }

func (endingBlankLineRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		if len(file.Lines) == 0 {
			continue
		}

		last := file.Lines[len(file.Lines)-1]
		if !file.EndsWithNewline || last.IsBlank {
			findings = append(findings, finding(
				endingBlankLineRule{}.ID(),
				lint.SeverityWarn,
				file.Path,
				last.Number,
				"file must end with exactly one final newline",
			))
		}
	}
	return findings, nil
}
