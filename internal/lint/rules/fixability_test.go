/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package rules_test pins rule fixability against what the shared --fix pass
// actually does. Fixability is intrinsic: a rule is fixable exactly when it
// carries a Fix method (lint.FixableRule). Two properties keep the fix pipeline
// honest:
//
//   - the per-rule drift test proves each rule's Fix eliminates its own finding
//     and that non-fixable rules survive the full composition untouched, so the
//     empirical fixable set stays exactly the five formatting rules, and
//   - the equivalence test proves composing every rule's fix reproduces the
//     original whole-file Normalize byte for byte, against a frozen copy of it,
//     so decomposing Normalize into per-rule transforms changed no behavior.
package rules_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/envfile"
	"github.com/envaar/vaar/internal/lint"
	"github.com/envaar/vaar/internal/lint/rules"
)

// expectedFixable is the empirical set of rules the --fix pass repairs. The
// drift test fails if any rule enters or leaves this set, which is a real
// finding about the decomposition rather than a value to blindly update.
var expectedFixable = map[string]bool{
	"bom-character":       true,
	"line-ending":         true,
	"trailing-whitespace": true,
	"extra-blank-line":    true,
	"ending-blank-line":   true,
}

// fixabilityFixtures maps each registered rule ID to a bad dotenv fixture that
// triggers exactly that rule, mirroring the per-rule bad inputs in
// internal/lint/rules/deterministic/*_test.go. A rule with no constructible bad
// fixture must be listed in fixabilityUnconstructible with a reason instead of
// left out silently; the coverage test fails on any registered rule that is
// neither mapped nor excused.
func fixabilityFixtures() map[string][]byte {
	return map[string][]byte{
		"bom-character":        append([]byte{0xEF, 0xBB, 0xBF}, []byte("KEY=value\n")...),
		"constant-case":        []byte("database_url=value\n"),
		"duplicate-key":        []byte("KEY=value\nKEY=other\n"),
		"ending-blank-line":    []byte("KEY=value\n\n"),
		"extra-blank-line":     []byte("KEY=value\n\n\n"),
		"incorrect-delimiter":  []byte("BAD: value\n"),
		"invalid-key-name":     []byte("1INVALID=value\n"),
		"key-without-value":    []byte("NO_VALUE\n"),
		"leading-character":    []byte(" KEY=value\n"),
		"line-ending":          []byte("KEY=one\r\nKEY2=two\n"),
		"quote-character":      []byte("KEY=\"broken\n"),
		"space-character":      []byte("KEY = value\n"),
		"substitution-key":     []byte("BROKEN=${BAR\n"),
		"trailing-whitespace":  []byte("KEY=value  \n"),
		"value-without-key":    []byte("=supersecret-token-123\n"),
		"value-without-quotes": []byte("FOO=BAR BAZ\n"),
	}
}

// fixabilityUnconstructible names any rule that genuinely has no bad fixture,
// with the reason it is excused. It is empty today because every built-in rule
// has a constructible trigger.
var fixabilityUnconstructible = map[string]string{}

// hasFinding reports whether a rule flags the given dotenv bytes.
func hasFinding(t *testing.T, rule lint.Rule, data []byte) bool {
	t.Helper()

	file, err := envfile.Parse("drift.env", data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	findings, err := rule.Run(lint.Context{Files: []envfile.File{file}})
	if err != nil {
		t.Fatalf("rule run failed: %v", err)
	}
	return len(findings) > 0
}

// TestFixabilityMatchesFixPass is the load-bearing drift test. For every
// registered rule it asserts, against that rule's bad fixture:
//
//   - a rule with a fix half eliminates its own finding when its Fix runs
//     alone and again under the full composition, and
//   - a rule without a fix half keeps its finding after the full composition.
func TestFixabilityMatchesFixPass(t *testing.T) {
	fixtures := fixabilityFixtures()
	all := rules.All()

	for _, rule := range all {
		rule := rule
		t.Run(rule.ID(), func(t *testing.T) {
			if reason, excused := fixabilityUnconstructible[rule.ID()]; excused {
				t.Skipf("no constructible bad fixture: %s", reason)
			}

			fixture, ok := fixtures[rule.ID()]
			if !ok {
				t.Fatalf("rule %q has no fixability fixture and is not excused in fixabilityUnconstructible", rule.ID())
			}

			// Guard the fixture: it must actually trigger the rule, or the drift
			// assertions below would be vacuous.
			if !hasFinding(t, rule, fixture) {
				t.Fatalf("fixture for %q does not trigger the rule; fix the fixture", rule.ID())
			}

			composed := !hasFinding(t, rule, lint.FixData(all, fixture))

			if lint.IsFixable(rule) {
				fixable, ok := rule.(lint.FixableRule)
				if !ok {
					t.Fatalf("rule %q reads as fixable but has no Fix method", rule.ID())
				}
				if hasFinding(t, rule, fixable.Fix(fixture)) {
					t.Fatalf("rule %q has a fix half but its own Fix does not eliminate its finding", rule.ID())
				}
				if !composed {
					t.Fatalf("rule %q has a fix half but the full composition leaves its finding; is it missing from the fix pipeline order?", rule.ID())
				}
				return
			}

			if composed {
				t.Fatalf("rule %q has no fix half yet the full composition eliminates its finding; add a Fix method or narrow a fix transform", rule.ID())
			}
		})
	}
}

// TestEmpiricalFixableSet asserts the set of fixable rules is exactly the five
// formatting rules the fix pass owns. A mismatch is a real finding about the
// decomposition, not a number to update blindly.
func TestEmpiricalFixableSet(t *testing.T) {
	got := map[string]bool{}
	for _, rule := range rules.All() {
		if lint.IsFixable(rule) {
			got[rule.ID()] = true
		}
	}

	if len(got) != len(expectedFixable) {
		t.Fatalf("fixable set has %d rules, want %d: got %v", len(got), len(expectedFixable), sortedKeys(got))
	}
	for id := range expectedFixable {
		if !got[id] {
			t.Fatalf("expected rule %q to be fixable, but it is not", id)
		}
	}
	for id := range got {
		if !expectedFixable[id] {
			t.Fatalf("rule %q is fixable but not in the expected set; is that intended?", id)
		}
	}
}

// TestEveryRuleHasFixabilityCoverage fails if a newly registered rule lacks
// both a fixture and a named exclusion, so the drift test can never silently
// skip a rule.
func TestEveryRuleHasFixabilityCoverage(t *testing.T) {
	fixtures := fixabilityFixtures()
	for _, rule := range rules.All() {
		_, hasFixture := fixtures[rule.ID()]
		_, excused := fixabilityUnconstructible[rule.ID()]
		if !hasFixture && !excused {
			t.Errorf("rule %q has neither a fixability fixture nor a named exclusion", rule.ID())
		}
	}
}

// TestFixDataMatchesLegacyNormalize pins the load-bearing invariant: composing
// every rule's fix half reproduces the original whole-file Normalize exactly,
// EXCEPT where a fix half is now deliberately scoped to its own finding and so
// intentionally diverges from the blunt legacy pass. legacyNormalize is a frozen
// copy of the pre-decomposition implementation, so the comparison is against real
// historical behavior, not a reimplementation.
//
// The divergences are the two maintainer-approved scope corrections:
//
//   - line-ending: its fix now runs only when the file actually MIXES CRLF and
//     LF (the same condition the rule reports). A file with uniform endings — all
//     CRLF or all lone CR — has no line-ending finding, so a scoped --fix leaves
//     its endings untouched instead of blindly forcing LF the way legacy did.
//   - trailing-whitespace: its fix now empties a line only when the line is
//     nothing but spaces and tabs (what the rule models). A line whose only
//     "blank" byte is a vertical tab or form feed is not a trailing-whitespace
//     finding, so it is kept (trailing spaces/tabs trimmed) rather than emptied.
//
// The affected corpus cases therefore assert the new finding-scoped bytes rather
// than the legacy LF-forcing / whitespace-emptying output.
func TestFixDataMatchesLegacyNormalize(t *testing.T) {
	all := rules.All()
	diverged := map[string][]byte{
		// line-ending scoped to mixed files: uniform endings pass through.
		"crlf":              []byte("A=1\r\nB=2\r\n"),
		"lone-cr":           []byte("A=1\rB=2\r"),
		"tabs-and-crlf":     []byte("KEY=value\r\n\r\nNEXT=x\r\n"),
		"bom-crlf-trailing": []byte("KEY=value\r\n\r\nFOO=bar\r\n"),
		// trailing-whitespace scoped to spaces/tabs: vertical-tab and form-feed
		// lines keep their non-space/tab content instead of being emptied.
		"vertical-tab-blank": []byte("A=1\n  \v\nB=2\n"),
		"form-feed-blank":    []byte("A=1\n\f\nB=2\n"),
	}
	for _, tc := range equivalenceCorpus(t) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			want := legacyNormalize(tc.data)
			if override, ok := diverged[tc.name]; ok {
				// This case intentionally diverges from legacy because a fix half
				// is now scoped to its finding; assert the new correct bytes.
				want = override
			}
			got := lint.FixData(all, tc.data)
			if string(got) != string(want) {
				t.Fatalf("FixData disagrees with expected on %q\ninput %q\ngot   %q\nwant  %q",
					tc.name, tc.data, got, want)
			}
		})
	}
}

type corpusCase struct {
	name string
	data []byte
}

// equivalenceCorpus is the repo's own testdata dotenv files plus a battery of
// synthetic nasty inputs the fix pass must handle identically to Normalize.
func equivalenceCorpus(t *testing.T) []corpusCase {
	t.Helper()

	cases := []corpusCase{
		{"empty", []byte("")},
		{"single-newline", []byte("\n")},
		{"all-blank", []byte("\n\n\n")},
		{"whitespace-only-spaces", []byte("   \n\t\n  \t \n")},
		{"whitespace-only-no-newline", []byte("   ")},
		{"vertical-tab-blank", []byte("A=1\n  \v  \nB=2\n")},
		{"form-feed-blank", []byte("A=1\n\f\nB=2\n")},
		{"no-final-newline", []byte("KEY=value")},
		{"trailing-spaces", []byte("KEY=value   \n")},
		{"trailing-tabs", []byte("KEY=value\t\t\n")},
		{"double-blank", []byte("A=1\n\n\nB=2\n")},
		{"trailing-blank-lines", []byte("KEY=value\n\n\n")},
		{"crlf", []byte("A=1\r\nB=2\r\n")},
		{"lone-cr", []byte("A=1\rB=2\r")},
		{"mixed-endings", []byte("A=1\r\nB=2\nC=3\r\n")},
		{"bom-only", []byte{0xEF, 0xBB, 0xBF}},
		{"bom-plus-key", append([]byte{0xEF, 0xBB, 0xBF}, []byte("KEY=value\n")...)},
		{"bom-crlf-trailing", append([]byte{0xEF, 0xBB, 0xBF}, []byte("KEY=value  \r\n\r\n\r\nFOO=bar   ")...)},
		{"bom-mid-file-untouched", []byte("A=1\n\ufeffB=2\n")},
		{"content-then-blanks-then-content", []byte("A=1  \n\n\n\nB=2\t\n\n")},
		{"comment-and-blanks", []byte("# note\n\n\n#end\n\n")},
		{"tabs-and-crlf", []byte("KEY=value\t\r\n\r\nNEXT=x\r\n")},
		{"leading-ws-preserved", []byte("  KEY=value  \n")},
		{"only-crlf-blanks", []byte("\r\n\r\n\r\n")},
	}

	root := filepath.Join("..", "..", "..", "testdata")
	matches, err := filepath.Glob(filepath.Join(root, "envfile", "*", "*.env"))
	if err != nil {
		t.Fatalf("glob testdata failed: %v", err)
	}
	sort.Strings(matches)
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s failed: %v", path, err)
		}
		cases = append(cases, corpusCase{name: "testdata/" + filepath.Base(path), data: data})
	}

	return cases
}

// legacyNormalize is a frozen, verbatim copy of the original
// envfile.Normalize implementation that the per-rule fix halves replaced. It
// exists only here so TestFixDataMatchesLegacyNormalize can pin the new
// composition against real historical behavior.
func legacyNormalize(data []byte) []byte {
	lines := envfile.Split(data)
	if len(lines) == 0 {
		return []byte{}
	}

	normalized := make([]string, 0, len(lines))
	prevBlank := false

	for i, raw := range lines {
		content := raw.Content
		if i == 0 && strings.HasPrefix(content, "\ufeff") {
			content = strings.TrimPrefix(content, "\ufeff")
		}

		content = strings.TrimRight(content, " \t")
		blank := strings.TrimSpace(content) == ""
		if blank {
			if prevBlank {
				continue
			}
			prevBlank = true
			normalized = append(normalized, "")
			continue
		}

		prevBlank = false
		normalized = append(normalized, content)
	}

	for len(normalized) > 0 && normalized[len(normalized)-1] == "" {
		normalized = normalized[:len(normalized)-1]
	}

	if len(normalized) == 0 {
		return []byte{}
	}

	var builder strings.Builder
	for _, line := range normalized {
		builder.WriteString(line)
		builder.WriteByte('\n')
	}

	return []byte(builder.String())
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
