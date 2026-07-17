/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint_test

import (
	"bytes"
	"testing"

	"github.com/envaar/vaar/internal/lint"
)

// plainRule implements Rule without a fix half.
type plainRule struct{}

func (plainRule) ID() string                               { return "plain" }
func (plainRule) Description() string                      { return "no fix half" }
func (plainRule) Run(lint.Context) ([]lint.Finding, error) { return nil, nil }

// fixableRule implements Rule and supplies a fix half, so it is intrinsically
// fixable: there is no separate boolean that could disagree with the method.
type fixableRule struct{ plainRule }

func (fixableRule) ID() string { return "line-ending" }
func (fixableRule) Fix(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

func TestIsFixable(t *testing.T) {
	cases := []struct {
		name string
		rule lint.Rule
		want bool
	}{
		{name: "no fix half", rule: plainRule{}, want: false},
		{name: "has fix half", rule: fixableRule{}, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := lint.IsFixable(tc.rule); got != tc.want {
				t.Fatalf("IsFixable(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

// TestFixDataAppliesOnlyProvidedRuleFixes proves the composition runs a rule's
// fix half only when that rule is passed in, and applies nothing for a rule
// without a fix half.
func TestFixDataAppliesOnlyProvidedRuleFixes(t *testing.T) {
	input := []byte("KEY=one\r\nKEY2=two\r\n")

	if got := lint.FixData([]lint.Rule{plainRule{}}, input); !bytes.Equal(got, input) {
		t.Fatalf("a rule without a fix half must not change data: got %q", got)
	}

	want := []byte("KEY=one\nKEY2=two\n")
	if got := lint.FixData([]lint.Rule{fixableRule{}}, input); !bytes.Equal(got, want) {
		t.Fatalf("line-ending fix half not applied: got %q want %q", got, want)
	}
}
