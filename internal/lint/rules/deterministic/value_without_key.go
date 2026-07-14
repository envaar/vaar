/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
)

type valueWithoutKeyRule struct{}

// NewValueWithoutKey returns a rule that flags values or delimiters that
// appear without a key.
func NewValueWithoutKey() lint.Rule { return valueWithoutKeyRule{} }

func (valueWithoutKeyRule) ID() string { return "value-without-key" }

func (valueWithoutKeyRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if line.HasKey || line.IsBlank || line.IsComment {
				continue
			}
			if line.DelimiterState == envfile.DelimiterEquals || line.DelimiterState == envfile.DelimiterColon || line.Value != "" {
				findings = append(findings, finding(
					valueWithoutKeyRule{}.ID(),
					lint.SeverityError,
					file.Path,
					line.Number,
					"value appears without a valid key",
				))
			}
		}
	}
	return findings, nil
}
