/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestDuplicateKey(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewDuplicateKey(),
		input:           []byte("KEY=value\nKEY=other\n"),
		wantCount:       1,
		wantLine:        2,
		wantRule:        "duplicate-key",
		wantSeverity:    "error",
		wantMessagePart: "KEY is defined more than once",
	})
}
