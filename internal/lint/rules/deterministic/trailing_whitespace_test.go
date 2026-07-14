/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestTrailingWhitespace(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewTrailingWhitespace(),
		input:           []byte("KEY=value  \n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "trailing-whitespace",
		wantSeverity:    "warn",
		wantMessagePart: "trailing whitespace",
	})
}
