/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"github.com/envaar/vaar/internal/analysis"
	"github.com/envaar/vaar/internal/lint"
)

type valueWithoutQuotesRule struct{}

// NewValueWithoutQuotes returns a rule that flags unquoted values
// containing whitespace.
func NewValueWithoutQuotes() lint.Rule { return valueWithoutQuotesRule{} }

func (valueWithoutQuotesRule) ID() string { return "value-without-quotes" }

func (valueWithoutQuotesRule) Description() string {
	return "flags unquoted values containing whitespace"
}

func (valueWithoutQuotesRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)

	ctx.Snapshot.RangeDocuments(func(document analysis.DocumentView) {
		document.RangeLines(func(line analysis.Line) {
			if !line.HasAssignment || !line.HasValue {
				return
			}

			if line.QuoteState != analysis.QuoteNone {
				return
			}

			if line.ValueContainsWhitespace {
				findings = append(findings, finding(
					valueWithoutQuotesRule{}.ID(),
					lint.SeverityError,
					document.DisplayPath,
					line.Number,
					"value containing whitespace should be enclosed in quotes",
				))
			}
		})
	})

	return findings, nil
}
