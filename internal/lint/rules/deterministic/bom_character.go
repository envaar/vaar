/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
)

type bomCharacterRule struct{}

// NewBOMCharacter returns a rule that flags a UTF-8 BOM at the start of a
// dotenv file, because the marker changes the first key on disk.
func NewBOMCharacter() lint.Rule { return bomCharacterRule{} }

func (bomCharacterRule) ID() string { return "bom-character" }
func (bomCharacterRule) Description() string {
	return "flags a UTF-8 BOM at the start of a dotenv file"
}

// Fix strips the leading UTF-8 BOM so the first key on disk is not altered.
func (bomCharacterRule) Fix(data []byte) []byte { return envfile.StripBOM(data) }

func (bomCharacterRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)
	for _, file := range ctx.Files {
		if !file.BOM {
			continue
		}
		lineNumber := 1
		if len(file.Lines) > 0 {
			lineNumber = file.Lines[0].Number
		}
		findings = append(findings, finding(
			bomCharacterRule{}.ID(),
			lint.SeverityWarn,
			file.Path,
			lineNumber,
			"file starts with a UTF-8 BOM",
		))
	}
	return findings, nil
}
