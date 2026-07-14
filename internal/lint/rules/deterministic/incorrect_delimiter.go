/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"fmt"

	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
)

type incorrectDelimiterRule struct{}

// NewIncorrectDelimiter returns a rule that flags `:` where dotenv assignments
// should use `=`.
func NewIncorrectDelimiter() lint.Rule { return incorrectDelimiterRule{} }

func (incorrectDelimiterRule) ID() string { return "incorrect-delimiter" }

func (incorrectDelimiterRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if !line.HasKey || line.DelimiterState != envfile.DelimiterColon {
				continue
			}
			findings = append(findings, finding(
				incorrectDelimiterRule{}.ID(),
				lint.SeverityError,
				file.Path,
				line.Number,
				fmt.Sprintf("%s uses ':' instead of '='", line.Key),
			))
		}
	}
	return findings, nil
}
