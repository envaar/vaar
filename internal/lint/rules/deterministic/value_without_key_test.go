/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestValueWithoutKey(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:               deterministic.NewValueWithoutKey(),
		input:              []byte("=supersecret-token-123\n"),
		wantCount:          1,
		wantLine:           1,
		wantRule:           "value-without-key",
		wantSeverity:       "error",
		wantMessagePart:    "value appears without a valid key",
		wantNotMessagePart: "supersecret-token-123",
	})
}
