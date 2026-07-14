/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestLeadingCharacter(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewLeadingCharacter(),
		input:           []byte(" KEY=value\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "leading-character",
		wantSeverity:    "warn",
		wantMessagePart: "leading whitespace",
	})
}
