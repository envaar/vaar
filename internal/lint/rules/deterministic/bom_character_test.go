/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestBOMCharacter(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewBOMCharacter(),
		input:           append([]byte{0xEF, 0xBB, 0xBF}, []byte("KEY=value\n")...),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "bom-character",
		wantSeverity:    "warn",
		wantMessagePart: "UTF-8 BOM",
	})
}
