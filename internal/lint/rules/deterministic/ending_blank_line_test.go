/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestEndingBlankLine(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewEndingBlankLine(),
		input:           []byte("KEY=value\n\n"),
		wantCount:       1,
		wantLine:        2,
		wantRule:        "ending-blank-line",
		wantSeverity:    "warn",
		wantMessagePart: "exactly one final newline",
	})
}
