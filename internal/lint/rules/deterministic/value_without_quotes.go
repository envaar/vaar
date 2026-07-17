/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"strings"

	"github.com/envaar/vaar/internal/envfile"
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

	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if !line.HasAssignment || !line.HasValue {
				continue
			}

			if line.Value == "" {
				continue
			}

			if line.QuoteState != envfile.QuoteNone {
				continue
			}

			if strings.ContainsRune(line.Value, ' ') ||
				strings.ContainsRune(line.Value, '\t') {
				findings = append(findings, finding(
					valueWithoutQuotesRule{}.ID(),
					lint.SeverityError,
					file.Path,
					line.Number,
					"value containing whitespace should be enclosed in quotes",
				))
			}
		}
	}

	return findings, nil
}
