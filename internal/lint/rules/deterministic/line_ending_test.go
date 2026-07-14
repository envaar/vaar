/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestLineEnding(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewLineEnding(),
		input:           []byte("KEY=one\r\nKEY2=two\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "line-ending",
		wantSeverity:    "warn",
		wantMessagePart: "mixed CRLF and LF line endings",
	})
}
