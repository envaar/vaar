/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestExtraBlankLine(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewExtraBlankLine(),
		input:           []byte("KEY=value\n\n\n"),
		wantCount:       1,
		wantLine:        3,
		wantRule:        "extra-blank-line",
		wantSeverity:    "warn",
		wantMessagePart: "repeated blank line",
	})
}
