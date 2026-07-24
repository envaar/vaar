/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic

import (
	"fmt"
	"strings"

	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
)

type substitutionKeyRule struct{}

// NewSubstitutionKey returns a rule that flags malformed env-var
// substitutions inside assignment values.
func NewSubstitutionKey() lint.Rule { return substitutionKeyRule{} }

func (substitutionKeyRule) ID() string { return "substitution-key" }

func (substitutionKeyRule) Description() string {
	return "flags malformed env-var substitution syntax inside values"
}

func (substitutionKeyRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	findings := make([]lint.Finding, 0)

	for _, file := range ctx.Files {
		for _, line := range file.Lines {
			if !line.HasAssignment || !line.HasValue {
				continue
			}
			if line.QuoteState == envfile.QuoteSingle ||
				line.QuoteState == envfile.QuoteUnbalanced {
				// Single-quoted values are literal text, and quote-character owns
				// malformed wrapping quotes, so this rule stays focused on
				// substitution syntax in values where substitution-like text is
				// semantically active.
				continue
			}

			findings = append(findings, scanSubstitutionFindings(
				file.Path,
				line.Number,
				line.Value,
			)...)
		}
	}

	return findings, nil
}

func scanSubstitutionFindings(path string, line int, value string) []lint.Finding {
	findings := make([]lint.Finding, 0)

	for i := 0; i < len(value); {
		if value[i] != '$' {
			i++
			continue
		}

		if i+1 >= len(value) {
			i++
			continue
		}

		if value[i+1] == '{' {
			finding, next := scanBracedSubstitution(path, line, value, i)
			if finding != nil {
				findings = append(findings, *finding)
			}
			i = next
			continue
		}

		if !isPortableKeyStart(value[i+1]) {
			i++
			continue
		}

		end := i + 2
		for end < len(value) && isPortableKeyContinue(value[end]) {
			end++
		}

		key := value[i+1 : end]
		if !validKeyName(key) {
			i = end
			continue
		}

		if end < len(value) && value[end] == '}' {
			tokenEnd := end + 1
			for tokenEnd < len(value) && value[tokenEnd] == '}' {
				tokenEnd++
			}
			findings = append(findings, finding(
				substitutionKeyRule{}.ID(),
				lint.SeverityError,
				path,
				line,
				fmt.Sprintf(`substitution %q contains an unmatched closing "}"`, value[i:tokenEnd]),
			))
			i = tokenEnd
			continue
		}

		i = end
	}

	return findings
}

func scanBracedSubstitution(path string, line int, value string, start int) (*lint.Finding, int) {
	closeOffset := strings.IndexByte(value[start+2:], '}')
	if closeOffset < 0 {
		finding := finding(
			substitutionKeyRule{}.ID(),
			lint.SeverityError,
			path,
			line,
			fmt.Sprintf(`substitution %q is missing a closing "}"`, value[start:]),
		)
		return &finding, len(value)
	}

	closeIdx := start + 2 + closeOffset
	token := value[start : closeIdx+1]
	key := value[start+2 : closeIdx]

	if key == "" {
		finding := finding(
			substitutionKeyRule{}.ID(),
			lint.SeverityError,
			path,
			line,
			fmt.Sprintf("substitution %q is empty", token),
		)
		return &finding, closeIdx + 1
	}

	if !validKeyName(key) {
		finding := finding(
			substitutionKeyRule{}.ID(),
			lint.SeverityError,
			path,
			line,
			fmt.Sprintf("substitution %q does not use a portable env key name", token),
		)
		return &finding, closeIdx + 1
	}

	if closeIdx+1 < len(value) && value[closeIdx+1] == '}' {
		tokenEnd := closeIdx + 2
		for tokenEnd < len(value) && value[tokenEnd] == '}' {
			tokenEnd++
		}
		finding := finding(
			substitutionKeyRule{}.ID(),
			lint.SeverityError,
			path,
			line,
			fmt.Sprintf(`substitution %q contains an unmatched closing "}"`, value[start:tokenEnd]),
		)
		return &finding, tokenEnd
	}

	return nil, closeIdx + 1
}

func isPortableKeyStart(b byte) bool {
	return b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func isPortableKeyContinue(b byte) bool {
	return isPortableKeyStart(b) || (b >= '0' && b <= '9')
}
