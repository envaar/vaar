/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package report_test verifies the human and machine report formats.
package report_test

import (
	"encoding/json"
	"testing"

	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/report"
)

func TestTextRender(t *testing.T) {
	findings := []lint.Finding{{
		Rule:     "duplicate-key",
		Severity: lint.SeverityError,
		File:     ".env",
		Line:     4,
		Message:  "DATABASE_URL is defined more than once",
	}}

	got := report.Text(findings)
	want := "error duplicate-key .env:4 DATABASE_URL is defined more than once\n"
	if got != want {
		t.Fatalf("unexpected text output: got %q want %q", got, want)
	}
}

func TestTextRenderMarksFixedFindings(t *testing.T) {
	findings := []lint.Finding{{
		Rule:     "trailing-whitespace",
		Severity: lint.SeverityWarn,
		File:     ".env",
		Line:     1,
		Message:  "line has trailing whitespace",
		Fixed:    true,
	}}

	got := report.Text(findings)
	want := "[fixed] warn trailing-whitespace .env:1 line has trailing whitespace\n"
	if got != want {
		t.Fatalf("unexpected fixed text output: got %q want %q", got, want)
	}
}

func TestJSONRender(t *testing.T) {
	findings := []lint.Finding{{
		Rule:     "duplicate-key",
		Severity: lint.SeverityError,
		File:     ".env",
		Line:     4,
		Message:  "DATABASE_URL is defined more than once",
	}}

	payload, err := report.JSON(findings)
	if err != nil {
		t.Fatalf("json render failed: %v", err)
	}

	var decoded struct {
		Findings []lint.Finding `json:"findings"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if got, want := len(decoded.Findings), 1; got != want {
		t.Fatalf("unexpected finding count: got %d want %d", got, want)
	}
	if decoded.Findings[0].Rule != "duplicate-key" {
		t.Fatalf("unexpected rule in JSON payload: %#v", decoded.Findings[0])
	}
}
