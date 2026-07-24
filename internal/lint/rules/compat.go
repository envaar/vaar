/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package rules

import (
	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func NewDuplicateKey() lint.Rule { return deterministic.NewDuplicateKey() }

func NewIncorrectDelimiter() lint.Rule { return deterministic.NewIncorrectDelimiter() }

func NewKeyWithoutValue() lint.Rule { return deterministic.NewKeyWithoutValue() }

func NewValueWithoutKey() lint.Rule { return deterministic.NewValueWithoutKey() }

func NewLeadingCharacter() lint.Rule { return deterministic.NewLeadingCharacter() }

func NewQuoteCharacter() lint.Rule { return deterministic.NewQuoteCharacter() }

func NewValueWithoutQuotes() lint.Rule { return deterministic.NewValueWithoutQuotes() }

func NewSubstitutionKey() lint.Rule { return deterministic.NewSubstitutionKey() }

func NewSpaceCharacter() lint.Rule { return deterministic.NewSpaceCharacter() }

func NewTrailingWhitespace() lint.Rule { return deterministic.NewTrailingWhitespace() }

func NewEndingBlankLine() lint.Rule { return deterministic.NewEndingBlankLine() }

func NewExtraBlankLine() lint.Rule { return deterministic.NewExtraBlankLine() }

func NewInvalidKeyName() lint.Rule { return deterministic.NewInvalidKeyName() }

func NewConstantCase() lint.Rule { return deterministic.NewConstantCase() }

func NewBOMCharacter() lint.Rule { return deterministic.NewBOMCharacter() }

func NewLineEnding() lint.Rule { return deterministic.NewLineEnding() }
