/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"github.com/envaar/vaar/internal/analysis"
	"github.com/envaar/vaar/internal/lint"
)

type valueWithoutKeyRule struct{}

// NewValueWithoutKey returns a rule that flags values or delimiters that
// appear without a key.
func NewValueWithoutKey() lint.Rule { return valueWithoutKeyRule{} }

func (valueWithoutKeyRule) ID() string { return "value-without-key" }
func (valueWithoutKeyRule) Description() string {
	return "flags values or delimiters that appear without a key"
}

func (valueWithoutKeyRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, document := range ctx.Snapshot.Documents() {
		for _, line := range document.Lines {
			if line.HasKey || line.IsBlank || line.IsComment {
				continue
			}
			if line.DelimiterState == analysis.DelimiterEquals || line.DelimiterState == analysis.DelimiterColon || line.HasValue {
				findings = append(findings, finding(
					valueWithoutKeyRule{}.ID(),
					lint.SeverityError,
					document.DisplayPath,
					line.Number,
					"value appears without a valid key",
				))
			}
		}
	}
	return findings, nil
}
