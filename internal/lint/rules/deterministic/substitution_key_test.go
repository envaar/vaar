/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package deterministic_test

import (
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules/deterministic"
)

func TestSubstitutionKeyAcceptsValidForms(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule: deterministic.NewSubstitutionKey(),
		input: []byte(strings.Join([]string{
			"VALID_BRACED=${BAR}",
			"VALID_UNBRACED=$BAR",
			"VALID_QUOTED_BRACED=\"${BAR}\"",
			"VALID_QUOTED_UNBRACED=\"$BAR\"",
			"VALID_SINGLE_QUOTED='${BAR}-$BAZ'",
			"VALID_MIXED=\"prefix-${BAR}-$BAZ-suffix\"",
			"VALUE=$BAR # } is part of the comment",
			"LITERAL_BRACE=value}",
			"JSON_VALUE={\"enabled\":true}",
			"LITERAL_DOLLAR=$",
			"UNSUPPORTED_DEFAULT=${KEY:-fallback}",
			"UNSUPPORTED_REQUIRED=${KEY:?required}",
			"UNSUPPORTED_NESTED=${OUTER:-${INNER}}",
			"PLAIN_VALUE=value",
		}, "\n") + "\n"),
		wantCount: 0,
	})
}

func TestSubstitutionKeyReportsMalformedSubstitutions(t *testing.T) {
	findings := runSubstitutionKey(t, strings.Join([]string{
		"ABC=${BAR",
		"FOO=\"$BAR}\"",
		"EMPTY_REFERENCE=${}",
		"EXTRA_CLOSE=${BAR}}",
	}, "\n")+"\n")

	assertSubstitutionFindings(t, findings, []expectedFinding{
		{
			line:        1,
			messagePart: `substitution "${BAR" is missing a closing "}"`,
		},
		{
			line:        2,
			messagePart: `substitution "$BAR}" contains an unmatched closing "}"`,
		},
		{
			line:        3,
			messagePart: `substitution "${}" is empty`,
		},
		{
			line:        4,
			messagePart: `substitution "${BAR}}" contains an unmatched closing "}"`,
		},
	})
}

func TestSubstitutionKeySupportsMultipleSubstitutionsInOneValue(t *testing.T) {
	findings := runSubstitutionKey(t, "CHAIN=$FOO}$BAR}\n")

	assertSubstitutionFindings(t, findings, []expectedFinding{
		{
			line:        1,
			messagePart: `substitution "$FOO}" contains an unmatched closing "}"`,
		},
		{
			line:        1,
			messagePart: `substitution "$BAR}" contains an unmatched closing "}"`,
		},
	})
}

func TestSubstitutionKeySkipsSingleQuotedLiterals(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule: deterministic.NewSubstitutionKey(),
		input: []byte(strings.Join([]string{
			"LITERAL='${BAR'",
			"ALSO_LITERAL='$BAR}'",
			"EMPTY_LITERAL='${}'",
		}, "\n") + "\n"),
		wantCount: 0,
	})
}

func TestSubstitutionKeySkipsUnbalancedQuotedValues(t *testing.T) {
	runRuleTest(t, ruleTestCase{
		rule:      deterministic.NewSubstitutionKey(),
		input:     []byte("BROKEN=\"${BAR\n"),
		wantCount: 0,
	})
}

type expectedFinding struct {
	line        int
	messagePart string
}

// runSubstitutionKey parses input and runs only the substitution-key rule.
func runSubstitutionKey(t *testing.T, input string) []lint.Finding {
	t.Helper()

	file, err := envfile.Parse("test.env", []byte(input))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	findings, err := deterministic.NewSubstitutionKey().Run(lint.Context{
		Files: []envfile.File{file},
	})
	if err != nil {
		t.Fatalf("rule run failed: %v", err)
	}

	return findings
}

// assertSubstitutionFindings checks the stable fields this rule owns.
func assertSubstitutionFindings(t *testing.T, findings []lint.Finding, want []expectedFinding) {
	t.Helper()

	if got, wantCount := len(findings), len(want); got != wantCount {
		t.Fatalf("unexpected finding count: got %d want %d", got, wantCount)
	}

	for i, finding := range findings {
		if finding.Rule != "substitution-key" {
			t.Fatalf("finding %d rule: got %q want %q", i, finding.Rule, "substitution-key")
		}
		if finding.Severity != lint.SeverityError {
			t.Fatalf("finding %d severity: got %q want %q", i, finding.Severity, lint.SeverityError)
		}
		if finding.Line != want[i].line {
			t.Fatalf("finding %d line: got %d want %d", i, finding.Line, want[i].line)
		}
		if !strings.Contains(finding.Message, want[i].messagePart) {
			t.Fatalf("finding %d message: got %q want substring %q", i, finding.Message, want[i].messagePart)
		}
	}
}
