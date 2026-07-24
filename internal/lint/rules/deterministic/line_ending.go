/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
)

type lineEndingRule struct{}

// NewLineEnding returns a rule that flags files with mixed CRLF and LF
// endings because the same file should not switch newline style midstream.
func NewLineEnding() lint.Rule { return lineEndingRule{} }

func (lineEndingRule) ID() string          { return "line-ending" }
func (lineEndingRule) Description() string { return "flags files with mixed CRLF and LF line endings" }

// Fix converts CRLF and lone CR line endings to LF so the file uses one style,
// but only for a file that actually mixes CRLF and LF — the same condition Run
// uses to raise a finding. A file with uniform endings (all CRLF, all LF, or all
// lone CR) has no line-ending finding, so a scoped --fix leaves it byte for byte
// unchanged instead of blindly forcing LF on bytes this rule does not own.
func (lineEndingRule) Fix(data []byte) []byte {
	if !envfile.HasMixedLineEndings(data) {
		return data
	}
	return envfile.NormalizeLineEndings(data)
}

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
