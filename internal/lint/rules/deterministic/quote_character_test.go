/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestQuoteCharacter(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewQuoteCharacter(),
		input:           []byte("KEY=\"broken\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "quote-character",
		wantSeverity:    "error",
		wantMessagePart: "unbalanced quotes",
	})
}
