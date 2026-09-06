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

type duplicateKeyRule struct{}

// NewDuplicateKey returns a rule that flags the second and later assignment of
// the same key in one file.
func NewDuplicateKey() lint.Rule { return duplicateKeyRule{} }

func (duplicateKeyRule) ID() string { return "duplicate-key" }
func (duplicateKeyRule) Description() string {
	return "flags the second and later assignment of the same key in one file"
}

func (duplicateKeyRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, document := range ctx.Snapshot.Documents() {
		seen := make(map[string]int, len(document.Lines))
		for _, line := range document.Lines {
			if !line.HasKey || line.DelimiterState == analysis.DelimiterMissing {
				continue
			}
			if _, ok := seen[line.Key]; ok {
				findings = append(findings, finding(
					duplicateKeyRule{}.ID(),
					lint.SeverityError,
					document.DisplayPath,
					line.Number,
					fmt.Sprintf("%s is defined more than once", line.Key),
				))
				continue
			}
			seen[line.Key] = line.Number
		}
	}
	return findings, nil
}
