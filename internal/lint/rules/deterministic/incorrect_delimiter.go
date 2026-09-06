/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"fmt"

	"github.com/envaar/vaar/internal/analysis"
	"github.com/envaar/vaar/internal/lint"
)

type incorrectDelimiterRule struct{}

// NewIncorrectDelimiter returns a rule that flags `:` where dotenv assignments
// should use `=`.
func NewIncorrectDelimiter() lint.Rule { return incorrectDelimiterRule{} }

func (incorrectDelimiterRule) ID() string { return "incorrect-delimiter" }
func (incorrectDelimiterRule) Description() string {
	return "flags ':' where dotenv assignments should use '='"
}

func (incorrectDelimiterRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	ctx.Snapshot.RangeDocuments(func(document analysis.DocumentView) {
		document.RangeLines(func(line analysis.Line) {
			if !line.HasKey || line.DelimiterState != analysis.DelimiterColon {
				return
			}
			findings = append(findings, finding(
				incorrectDelimiterRule{}.ID(),
				lint.SeverityError,
				document.DisplayPath,
				line.Number,
				fmt.Sprintf("%s uses ':' instead of '='", line.Key),
			))
		})
	})
	return findings, nil
}
