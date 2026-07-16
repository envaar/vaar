/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"fmt"
	"strings"

	"github.com/envaar/vaar/internal/lint"
)

type constantCaseRule struct{}

// NewConstantCase returns a rule that warns about portable keys that are not
// CONSTANT_CASE.
func NewConstantCase() lint.Rule { return constantCaseRule{} }

func (constantCaseRule) ID() string { return "constant-case" }

func (constantCaseRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			// Structurally invalid keys belong to invalid-key-name; this
			// rule only reports otherwise-valid keys whose sole deviation
			// is lowercase letters.
			if !line.HasKey || !validKeyName(line.Key) {
				continue
			}
			upper := constantCaseKey(line.Key)
			if upper == line.Key {
				continue
			}
			findings = append(findings, finding(
				constantCaseRule{}.ID(),
				lint.SeverityWarn,
				file.Path,
				line.Number,
				fmt.Sprintf("%s should use CONSTANT_CASE: %s", line.Key, upper),
			))
		}
	}
	return findings, nil
}

// constantCaseKey uppercases ASCII lowercase letters and leaves every other
// rune untouched.
func constantCaseKey(key string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - ('a' - 'A')
		}
		return r
	}, key)
}
