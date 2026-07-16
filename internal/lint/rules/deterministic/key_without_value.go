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

type keyWithoutValueRule struct{}

// NewKeyWithoutValue returns a rule that flags bare keys and empty
// assignments that leave a variable unset.
func NewKeyWithoutValue() lint.Rule { return keyWithoutValueRule{} }

func (keyWithoutValueRule) ID() string { return "key-without-value" }
func (keyWithoutValueRule) Description() string {
	return "flags bare keys and empty assignments that leave a variable unset"
}

func (keyWithoutValueRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if !line.HasKey {
				continue
			}
			if line.DelimiterState == envfile.DelimiterMissing || (line.DelimiterState == envfile.DelimiterEquals && !line.HasValue) {
				message := fmt.Sprintf("%s is missing a value", line.Key)
				findings = append(findings, finding(
					keyWithoutValueRule{}.ID(),
					lint.SeverityError,
					file.Path,
					line.Number,
					message,
				))
			}
		}
	}
	return findings, nil
}
