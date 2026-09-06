/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/analysis"
	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
)

type engineTestRule struct {
	id       string
	calls    *int
	observed *[]string
	onRun    func()
	findings []lint.Finding
	err      error
}

func (r engineTestRule) ID() string          { return r.id }
func (r engineTestRule) Description() string { return "engine test rule" }

func (r engineTestRule) Run(ctx lint.Context) ([]lint.Finding, error) {
	if r.calls != nil {
		(*r.calls)++
	}
	if r.observed != nil {
		for _, document := range ctx.Snapshot.Documents() {
			*r.observed = append(*r.observed, document.DisplayPath)
		}
	}
	if r.onRun != nil {
		r.onRun()
	}
	return r.findings, r.err
}

func emptyAnalysisSnapshot() analysis.Snapshot {
	return analysis.NewSnapshot(nil)
}

func TestEngineRunsRulesAgainstAnalysisSnapshotAndSortsFindings(t *testing.T) {
	calls := 0
	var observed []string
	rule := engineTestRule{
		id:       "snapshot-rule",
		calls:    &calls,
		observed: &observed,
		findings: []lint.Finding{
			{Rule: "snapshot-rule", Severity: lint.SeverityWarn, File: "b.env", Line: 4, Message: "later"},
			{Rule: "snapshot-rule", Severity: lint.SeverityError, File: "a.env", Line: 2, Message: "first"},
		},
	}
	snapshot := analysis.NewSnapshot([]analysis.Document{
		{DisplayPath: "b.env", Lines: []analysis.Line{{Number: 4, Key: "B"}}},
		{DisplayPath: "a.env", Lines: []analysis.Line{{Number: 2, Key: "A"}}},
	})

	findings, err := lint.NewEngine(rule).Run(context.Background(), snapshot, lint.EngineOptions{})
	if err != nil {
		t.Fatalf("engine run failed: %v", err)
	}
	if calls != 1 {
		t.Fatalf("rule calls = %d, want 1", calls)
	}
	if want := []string{"b.env", "a.env"}; !reflect.DeepEqual(observed, want) {
		t.Fatalf("observed documents = %#v, want %#v", observed, want)
	}

	if got, want := len(findings), 2; got != want {
		t.Fatalf("finding count = %d, want %d", got, want)
	}
	if findings[0].File != "a.env" || findings[1].File != "b.env" {
		t.Fatalf("findings were not sorted deterministically: %#v", findings)
	}
}

func TestEngineRunsBuiltInRulesAgainstInMemoryAnalysis(t *testing.T) {
	snapshot := analysis.NewSnapshot([]analysis.Document{{
		DisplayPath:      ".env",
		BOM:              true,
		MixedLineEndings: true,
		EndsWithNewline:  true,
		Lines: []analysis.Line{
			{
				Number:                  1,
				Key:                     "lower",
				HasKey:                  true,
				HasValue:                true,
				HasAssignment:           true,
				QuoteState:              analysis.QuoteNone,
				SpaceBeforeDelimiter:    true,
				SpaceAfterDelimiter:     true,
				HasLeadingWhitespace:    true,
				HasTrailingWhitespace:   true,
				ValueContainsWhitespace: true,
			},
			{
				Number:         2,
				Key:            "lower",
				HasKey:         true,
				HasAssignment:  true,
				DelimiterState: analysis.DelimiterEquals,
			},
			{
				Number:         3,
				Key:            "1BAD",
				HasKey:         true,
				HasValue:       true,
				HasAssignment:  true,
				DelimiterState: analysis.DelimiterEquals,
			},
			{
				Number:         4,
				Key:            "COLON",
				HasKey:         true,
				HasValue:       true,
				DelimiterState: analysis.DelimiterColon,
			},
			{
				Number:         5,
				Key:            "NO_VALUE",
				HasKey:         true,
				DelimiterState: analysis.DelimiterMissing,
			},
			{
				Number:         6,
				HasValue:       true,
				DelimiterState: analysis.DelimiterEquals,
			},
			{
				Number:     7,
				QuoteState: analysis.QuoteUnbalanced,
			},
			{Number: 8, IsBlank: true},
			{Number: 9, IsBlank: true},
		},
	}})

	findings, err := lint.NewEngine(rules.All()...).Run(
		context.Background(), snapshot, lint.EngineOptions{},
	)
	if err != nil {
		t.Fatalf("engine run failed: %v", err)
	}

	gotRules := make(map[string]bool, len(findings))
	for _, finding := range findings {
		gotRules[finding.Rule] = true
		if finding.File != ".env" {
			t.Fatalf("finding file = %q, want .env", finding.File)
		}
	}

	wantRules := []string{
		"bom-character",
		"constant-case",
		"duplicate-key",
		"ending-blank-line",
		"extra-blank-line",
		"incorrect-delimiter",
		"invalid-key-name",
		"key-without-value",
		"leading-character",
		"line-ending",
		"quote-character",
		"space-character",
		"trailing-whitespace",
		"value-without-key",
		"value-without-quotes",
	}
	for _, ruleID := range wantRules {
		if !gotRules[ruleID] {
			t.Errorf("missing finding from built-in rule %q: %#v", ruleID, findings)
		}
	}
}

func TestEngineSelectionPreservesOnlyAndSkipSemantics(t *testing.T) {
	cases := []struct {
		name        string
		only        []string
		skip        []string
		wantCalls   map[string]int
		wantErrPart string
	}{
		{
			name:      "default selects all in declaration order",
			wantCalls: map[string]int{"first": 1, "second": 1},
		},
		{
			name:      "only selects named rules",
			only:      []string{"second"},
			wantCalls: map[string]int{"first": 0, "second": 1},
		},
		{
			name:      "skip removes named rules",
			skip:      []string{"first"},
			wantCalls: map[string]int{"first": 0, "second": 1},
		},
		{
			name:        "only and skip can leave no rules",
			only:        []string{"first"},
			skip:        []string{"first"},
			wantCalls:   map[string]int{"first": 0, "second": 0},
			wantErrPart: "no lint rules selected after applying --only and --skip",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			firstCalls, secondCalls := 0, 0
			engine := lint.NewEngine(
				engineTestRule{id: "first", calls: &firstCalls},
				engineTestRule{id: "second", calls: &secondCalls},
			)

			findings, err := engine.Run(context.Background(), emptyAnalysisSnapshot(), lint.EngineOptions{
				OnlyRules: tc.only,
				SkipRules: tc.skip,
			})
			if tc.wantErrPart != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrPart) {
					t.Fatalf("error = %v, want substring %q", err, tc.wantErrPart)
				}
				if findings != nil {
					t.Fatalf("findings = %#v, want nil on selection error", findings)
				}
				return
			}
			if err != nil {
				t.Fatalf("engine run failed: %v", err)
			}
			if findings == nil {
				t.Fatal("clean engine run returned nil findings")
			}
			if got, want := firstCalls, tc.wantCalls["first"]; got != want {
				t.Fatalf("first calls = %d, want %d", got, want)
			}
			if got, want := secondCalls, tc.wantCalls["second"]; got != want {
				t.Fatalf("second calls = %d, want %d", got, want)
			}
		})
	}
}

func TestEnginePreservesEmptyRuleSetSelectionBehavior(t *testing.T) {
	cases := []struct {
		name string
		only []string
		skip []string
	}{
		{name: "valid empty selection"},
		{name: "unknown only", only: []string{"missing"}},
		{name: "unknown skip", skip: []string{"missing"}},
		{name: "empty only", only: []string{""}},
		{name: "empty skip", skip: []string{""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings, err := lint.NewEngine().Run(context.Background(), emptyAnalysisSnapshot(), lint.EngineOptions{
				OnlyRules: tc.only,
				SkipRules: tc.skip,
			})
			if err != nil {
				t.Fatalf("engine run failed: %v", err)
			}
			if findings == nil {
				t.Fatal("valid empty engine run returned nil findings")
			}
			if len(findings) != 0 {
				t.Fatalf("findings = %#v, want empty", findings)
			}
		})
	}
}

func TestEngineRejectsUnknownAndEmptyRuleIDs(t *testing.T) {
	engine := lint.NewEngine(engineTestRule{id: "known"})
	cases := []struct {
		name string
		only []string
		skip []string
		want string
	}{
		{name: "unknown only", only: []string{"missing"}, want: `unknown lint rule "missing"`},
		{name: "unknown skip", skip: []string{"missing"}, want: `unknown lint rule "missing"`},
		{name: "empty only", only: []string{""}, want: "invalid empty rule ID in --only"},
		{name: "empty skip", skip: []string{""}, want: "invalid empty rule ID in --skip"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := engine.Run(context.Background(), emptyAnalysisSnapshot(), lint.EngineOptions{
				OnlyRules: tc.only,
				SkipRules: tc.skip,
			})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestEngineChecksCancellationBetweenRules(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	secondCalls := 0
	engine := lint.NewEngine(
		engineTestRule{id: "first", onRun: cancel},
		engineTestRule{id: "second", calls: &secondCalls},
	)

	_, err := engine.Run(ctx, emptyAnalysisSnapshot(), lint.EngineOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if secondCalls != 0 {
		t.Fatalf("second rule calls = %d, want 0 after cancellation", secondCalls)
	}
}

func TestEngineWrapsRuleErrorsWithRuleID(t *testing.T) {
	sentinel := errors.New("sentinel rule failure")
	_, err := lint.NewEngine(engineTestRule{id: "broken-rule", err: sentinel}).Run(
		context.Background(),
		emptyAnalysisSnapshot(),
		lint.EngineOptions{},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want wrapped sentinel", err)
	}
	if !strings.Contains(err.Error(), "broken-rule: sentinel rule failure") {
		t.Fatalf("error = %v, want rule ID and message", err)
	}
}

func TestRuleContextExposesOnlyAnalysisSnapshot(t *testing.T) {
	typ := reflect.TypeOf(lint.Context{})
	snapshotType := reflect.TypeOf(analysis.Snapshot{})
	if typ.NumField() != 1 {
		t.Fatalf("lint.Context has %d fields, want exactly one", typ.NumField())
	}
	field := typ.Field(0)
	if field.Name != "Snapshot" || field.Type != snapshotType {
		t.Fatalf("lint.Context field = %s %s, want Snapshot %s", field.Name, field.Type, snapshotType)
	}

	forbidden := []string{"Root", "Files", "Options", "Target", "TargetDir", "Fix", "Output"}
	for _, name := range forbidden {
		if _, ok := typ.FieldByName(name); ok {
			t.Fatalf("lint.Context exposes forbidden field %q", name)
		}
	}
}
