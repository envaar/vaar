/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestIncorrectDelimiter(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewIncorrectDelimiter(),
		input:           []byte("BAD: value\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "incorrect-delimiter",
		wantSeverity:    "error",
		wantMessagePart: "uses ':' instead of '='",
	})
}
