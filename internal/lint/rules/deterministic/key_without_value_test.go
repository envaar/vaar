/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestKeyWithoutValue(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewKeyWithoutValue(),
		input:           []byte("NO_VALUE\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "key-without-value",
		wantSeverity:    "error",
		wantMessagePart: "is missing a value",
	})
}
