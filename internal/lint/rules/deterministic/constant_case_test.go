/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"testing"

	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestConstantCaseFlagsLowercaseKey(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewConstantCase(),
		input:           []byte("database_url=value\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "constant-case",
		wantSeverity:    "warn",
		wantMessagePart: "database_url should use CONSTANT_CASE: DATABASE_URL",
	})
}

func TestConstantCaseFlagsMixedCaseKey(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewConstantCase(),
		input:           []byte("Database_URL=value\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "constant-case",
		wantSeverity:    "warn",
		wantMessagePart: "Database_URL should use CONSTANT_CASE: DATABASE_URL",
	})
}

func TestConstantCaseFlagsCamelCaseKey(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewConstantCase(),
		input:           []byte("apiToken=value\n"),
		wantCount:       1,
		wantLine:        1,
		wantRule:        "constant-case",
		wantSeverity:    "warn",
		wantMessagePart: "apiToken should use CONSTANT_CASE: APITOKEN",
	})
}

func TestConstantCaseAllowsUppercaseKeys(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:      deterministic.NewConstantCase(),
		input:     []byte("DATABASE_URL=value\nAPI_TOKEN=value\nAPI_V2_URL=value\n"),
		wantCount: 0,
	})
}

func TestConstantCaseAllowsDigitUnderscoreKey(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:      deterministic.NewConstantCase(),
		input:     []byte("_2=value\n"),
		wantCount: 0,
	})
}

// Structurally invalid keys are reported by invalid-key-name, not by
// constant-case: each rule stays in its own lane. Asserting only that
// constant-case stays silent would also pass if nothing flagged these keys at
// all, so pin both halves of the boundary: constant-case ignores them AND
// invalid-key-name reports them.
func TestConstantCaseLeavesInvalidKeysToInvalidKeyName(t *testing.T) {
	invalidKeys := []byte("api-key=value\n1api=value\n")

	runRuleTest(t, ruleTestCase{
		rule:      deterministic.NewConstantCase(),
		input:     invalidKeys,
		wantCount: 0,
	})

	runRuleTest(t, ruleTestCase{
		rule:            deterministic.NewInvalidKeyName(),
		input:           invalidKeys,
		wantCount:       2,
		wantLine:        1,
		wantRule:        "invalid-key-name",
		wantSeverity:    "error",
		wantMessagePart: "api-key is not a portable env key name",
	})
}
