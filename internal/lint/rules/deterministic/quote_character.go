/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"github.com/envaar/vaar/internal/analysis"
	"github.com/envaar/vaar/internal/lint"
)

type quoteCharacterRule struct{}

// NewQuoteCharacter returns a rule that flags values with unbalanced quotes or
// stray text after a closing quote.
func NewQuoteCharacter() lint.Rule { return quoteCharacterRule{} }

func (quoteCharacterRule) ID() string { return "quote-character" }
func (quoteCharacterRule) Description() string {
	return "flags values with unbalanced quotes or stray text after a closing quote"
}

func (quoteCharacterRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, document := range ctx.Snapshot.Documents() {
		for _, line := range document.Lines {
			if line.QuoteState != analysis.QuoteUnbalanced {
				continue
			}
			findings = append(findings, finding(
				quoteCharacterRule{}.ID(),
				lint.SeverityError,
				document.DisplayPath,
				line.Number,
				"value has unbalanced quotes",
			))
		}
	}
	return findings, nil
}
