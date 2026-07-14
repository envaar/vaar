/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import "github.com/envaar/vaar/internal/lint"

type spaceCharacterRule struct{}

// NewSpaceCharacter returns a rule that warns about spaces around the key,
// delimiter or value because they change the file's visual shape.
func NewSpaceCharacter() lint.Rule { return spaceCharacterRule{} }

func (spaceCharacterRule) ID() string { return "space-character" }

func (spaceCharacterRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if line.IsBlank || line.IsComment {
				continue
			}
			if !line.SpaceBeforeDelimiter && !line.SpaceAfterDelimiter {
				continue
			}
			findings = append(findings, finding(
				spaceCharacterRule{}.ID(),
				lint.SeverityWarn,
				file.Path,
				line.Number,
				"line has spaces around the key, delimiter or value",
			))
		}
	}
	return findings, nil
}
