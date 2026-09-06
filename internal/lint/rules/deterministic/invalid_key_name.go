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

type invalidKeyNameRule struct{}

// NewInvalidKeyName returns a rule that flags keys outside the portable
// env-var format.
func NewInvalidKeyName() lint.Rule { return invalidKeyNameRule{} }

func (invalidKeyNameRule) ID() string { return "invalid-key-name" }
func (invalidKeyNameRule) Description() string {
	return "flags keys outside the portable env-var format"
}

func (invalidKeyNameRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	ctx.Snapshot.RangeDocuments(func(document analysis.DocumentView) {
		document.RangeLines(func(line analysis.Line) {
			if !line.HasKey || validKeyName(line.Key) {
				return
			}
			findings = append(findings, finding(
				invalidKeyNameRule{}.ID(),
				lint.SeverityError,
				document.DisplayPath,
				line.Number,
				fmt.Sprintf("%s is not a portable env key name", line.Key),
			))
		})
	})
	return findings, nil
}
