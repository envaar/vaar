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

	for _, document := range ctx.Snapshot.Documents() {
		for _, line := range document.Lines {
			if !line.HasAssignment || !line.HasValue {
				continue
			}

			if line.QuoteState != analysis.QuoteNone {
				continue
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
		}
	}

	return findings, nil
}
