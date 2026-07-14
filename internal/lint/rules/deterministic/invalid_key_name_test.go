/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestInvalidKeyName(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewInvalidKeyName(),
		input:           []byte("1INVALID=value\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "invalid-key-name",
		wantSeverity:    "error",
		wantMessagePart: "not a portable env key name",
	})
}
