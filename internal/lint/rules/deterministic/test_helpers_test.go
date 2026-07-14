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
)

type ruleTestCase struct {
	rule               lint.Rule
	input              []byte
	wantCount          int
	wantLine           int
	wantRule           string
	wantSeverity       lint.Severity
	wantMessagePart    string
	wantNotMessagePart string
}

func runRuleTest(t *testing.T, tc ruleTestCase) {
	t.Helper()

	file, err := envfile.Parse("test.env", tc.input)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	findings, err := tc.rule.Run(lint.Context{Files: []envfile.File{file}})
	if err != nil {
		t.Fatalf("rule run failed: %v", err)
	}

	if got, want := len(findings), tc.wantCount; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
	if tc.wantCount == 0 {
		return
	}

	finding := findings[0]
	if finding.Rule != tc.wantRule {
		t.Fatalf("unexpected rule: got %q want %q", finding.Rule, tc.wantRule)
	}
	if finding.Severity != tc.wantSeverity {
		t.Fatalf("unexpected severity: got %q want %q", finding.Severity, tc.wantSeverity)
	}
	if finding.Line != tc.wantLine {
		t.Fatalf("unexpected line: got %d want %d", finding.Line, tc.wantLine)
	}
	if tc.wantMessagePart != "" && !strings.Contains(finding.Message, tc.wantMessagePart) {
		t.Fatalf("unexpected message: got %q want substring %q", finding.Message, tc.wantMessagePart)
	}
	if tc.wantNotMessagePart != "" && strings.Contains(finding.Message, tc.wantNotMessagePart) {
		t.Fatalf("message leaked sensitive value: %q", finding.Message)
	}
}
