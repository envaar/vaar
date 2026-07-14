/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"fmt"

	"github.com/envaar/vaar/internal/lint"
)

type invalidKeyNameRule struct{}

// NewInvalidKeyName returns a rule that flags keys outside the portable
// env-var format.
func NewInvalidKeyName() lint.Rule { return invalidKeyNameRule{} }

func (invalidKeyNameRule) ID() string { return "invalid-key-name" }

func (invalidKeyNameRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if !line.HasKey || validKeyName(line.Key) {
				continue
			}
			findings = append(findings, finding(
				invalidKeyNameRule{}.ID(),
				lint.SeverityError,
				file.Path,
				line.Number,
				fmt.Sprintf("%s is not a portable env key name", line.Key),
			))
		}
	}
	return findings, nil
}
