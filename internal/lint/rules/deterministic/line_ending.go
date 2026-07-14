/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import "github.com/envaar/vaar/internal/lint"

type lineEndingRule struct{}

// NewLineEnding returns a rule that flags files with mixed CRLF and LF
// endings because the same file should not switch newline style midstream.
func NewLineEnding() lint.Rule { return lineEndingRule{} }

func (lineEndingRule) ID() string { return "line-ending" }

func (lineEndingRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		if !file.MixedLineEndings {
			continue
		}
		lineNumber := 1
		if len(file.Lines) > 0 {
			lineNumber = file.Lines[0].Number
		}
		findings = append(findings, finding(
			lineEndingRule{}.ID(),
			lint.SeverityWarn,
			file.Path,
			lineNumber,
			"file uses mixed CRLF and LF line endings",
		))
	}
	return findings, nil
}
