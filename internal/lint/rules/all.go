/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package rules defines the built-in dotenv lint checks shipped with vaar and
// keeps a stable entrypoint over the category-specific rule packages.
package rules

import (
	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

// All returns the built-in rule set in the order the CLI and docs present it.
func All() []lint.Rule {
	return []lint.Rule{
		deterministic.NewDuplicateKey(),
		deterministic.NewIncorrectDelimiter(),
		deterministic.NewKeyWithoutValue(),
		deterministic.NewValueWithoutKey(),
		deterministic.NewLeadingCharacter(),
		deterministic.NewQuoteCharacter(),
		deterministic.NewSpaceCharacter(),
		deterministic.NewTrailingWhitespace(),
		deterministic.NewEndingBlankLine(),
		deterministic.NewExtraBlankLine(),
		deterministic.NewInvalidKeyName(),
		deterministic.NewBOMCharacter(),
		deterministic.NewLineEnding(),
	}
}
