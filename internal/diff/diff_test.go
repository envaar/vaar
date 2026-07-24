/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package diff

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/envaar/vaar/internal/envfile"
)

func TestCompareKeyPresence(t *testing.T) {
	tests := []struct {
		name                 string
		left                 string
		right                string
		wantMissingFromLeft  []string
		wantMissingFromRight []string
	}{
		{
			name:                 "identical key sets",
			left:                 "FOO=left\nBAR=left\n",
			right:                "BAR=right\nFOO=right\n",
			wantMissingFromLeft:  []string{},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "one key missing from left",
			left:                 "COMMON=left\n",
			right:                "COMMON=right\nBAR=right\n",
			wantMissingFromLeft:  []string{"BAR"},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "one key missing from right",
			left:                 "COMMON=left\nFOO=left\n",
			right:                "COMMON=right\n",
			wantMissingFromLeft:  []string{},
			wantMissingFromRight: []string{"FOO"},
		},
		{
			name:                 "keys missing from both files",
			left:                 "FOO=left\nCOMMON=left\n",
			right:                "BAR=right\nCOMMON=right\n",
			wantMissingFromLeft:  []string{"BAR"},
			wantMissingFromRight: []string{"FOO"},
		},
		{
			name:                 "multiple missing keys are sorted alphabetically",
			left:                 "DEBUG=true\nFOO=left\n",
			right:                "DATABASE_URL=right\nBAR=right\nBAZ=right\n",
			wantMissingFromLeft:  []string{"BAR", "BAZ", "DATABASE_URL"},
			wantMissingFromRight: []string{"DEBUG", "FOO"},
		},
		{
			name:                 "different values are treated as matching",
			left:                 "FOO=local-value\nCOMMON=local-common-value\n",
			right:                "FOO=example-value\nCOMMON=example-common-value\n",
			wantMissingFromLeft:  []string{},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "empty values count as present keys",
			left:                 "EMPTY=\nCOMMON=left\n",
			right:                "EMPTY=filled\nCOMMON=right\n",
			wantMissingFromLeft:  []string{},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "comments and blank lines are ignored",
			left:                 "\n  # comment\nFOO=left\n",
			right:                "FOO=right\n\n\t# another comment\nBAR=right\n",
			wantMissingFromLeft:  []string{"BAR"},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "inline comments do not affect key presence",
			left:                 "FOO=left # local comment\n",
			right:                "FOO=right # example comment\nBAR=right\n",
			wantMissingFromLeft:  []string{"BAR"},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "spacing around assignments does not affect key presence",
			left:                 " FOO = left\nCOMMON=left\n",
			right:                "FOO=right\n COMMON = right\nBAR=right\n",
			wantMissingFromLeft:  []string{"BAR"},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "duplicate declarations collapse into one key",
			left:                 "FOO=one\nFOO=two\nCOMMON=left\n",
			right:                "FOO=three\nCOMMON=right\nBAR=right\nBAR=again\n",
			wantMissingFromLeft:  []string{"BAR"},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "key comparison is case-sensitive",
			left:                 "FOO=left\n",
			right:                "foo=right\n",
			wantMissingFromLeft:  []string{"foo"},
			wantMissingFromRight: []string{"FOO"},
		},
		{
			name:                 "non-portable assignment keys are still compared",
			left:                 "api-key=left\nCOMMON=left\n",
			right:                "api-key=right\nCOMMON=right\n",
			wantMissingFromLeft:  []string{},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "quoted and unquoted values do not affect key presence",
			left:                 "FOO=\"quoted value\"\nBAR='single quoted'\n",
			right:                "FOO=unquoted\nBAR=plain\n",
			wantMissingFromLeft:  []string{},
			wantMissingFromRight: []string{},
		},
		{
			name:                 "empty file has no keys",
			left:                 "",
			right:                "FOO=right\nBAR=right\n",
			wantMissingFromLeft:  []string{"BAR", "FOO"},
			wantMissingFromRight: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Compare(".env", []byte(tc.left), ".env.example", []byte(tc.right))
			if err != nil {
				t.Fatalf("compare failed: %v", err)
			}

			assertResult(t, result, tc.wantMissingFromLeft, tc.wantMissingFromRight)
		})
	}
}

func TestCompareOnlyIncludesAssignmentKeys(t *testing.T) {
	result, err := Compare(
		".env",
		[]byte("FOO\nBAR:wrong-delimiter\nBAZ=value\n"),
		".env.example",
		[]byte("BAZ=example\n"),
	)
	if err != nil {
		t.Fatalf("compare failed: %v", err)
	}

	assertResult(t, result, []string{}, []string{})
}

func TestResultHasDifferences(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{
			name: "no missing keys",
			result: Result{
				MissingFromLeft:  []string{},
				MissingFromRight: []string{},
			},
			want: false,
		},
		{
			name: "missing from left",
			result: Result{
				MissingFromLeft: []string{"BAR"},
			},
			want: true,
		},
		{
			name: "missing from right",
			result: Result{
				MissingFromRight: []string{"FOO"},
			},
			want: true,
		},
		{
			name: "missing from both files",
			result: Result{
				MissingFromLeft:  []string{"BAR"},
				MissingFromRight: []string{"FOO"},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.result.HasDifferences(); got != tc.want {
				t.Fatalf("unexpected HasDifferences result: got %v want %v", got, tc.want)
			}
		})
	}
}

func TestCompareFilesReusesParsedFiles(t *testing.T) {
	left, err := envfile.Parse("left.env", []byte("FOO=left\n"))
	if err != nil {
		t.Fatalf("parse left failed: %v", err)
	}
	right, err := envfile.Parse("right.env", []byte("BAR=right\n"))
	if err != nil {
		t.Fatalf("parse right failed: %v", err)
	}

	result := CompareFiles(left, right)

	assertResult(t, result, []string{"BAR"}, []string{"FOO"})
	if result.Left != "left.env" {
		t.Fatalf("unexpected left path: got %q want %q", result.Left, "left.env")
	}
	if result.Right != "right.env" {
		t.Fatalf("unexpected right path: got %q want %q", result.Right, "right.env")
	}
}

func TestCompareResultDoesNotExposeValues(t *testing.T) {
	const value = "fake-password-value"

	result, err := Compare(
		".env",
		[]byte("DATABASE_PASSWORD="+value+"\n"),
		".env.example",
		[]byte("DATABASE_URL=fake-example-url\n"),
	)
	if err != nil {
		t.Fatalf("compare failed: %v", err)
	}

	rendered := fmt.Sprintf("%#v", result)
	if strings.Contains(rendered, value) {
		t.Fatalf("diff result leaked a value: %q", rendered)
	}
	if strings.Contains(rendered, "fake-example-url") {
		t.Fatalf("diff result leaked a value: %q", rendered)
	}
}

func TestCompareResultUsesPaths(t *testing.T) {
	result, err := Compare("left.env", []byte("FOO=left\n"), "right.env", []byte("BAR=right\n"))
	if err != nil {
		t.Fatalf("compare failed: %v", err)
	}

	if result.Left != "left.env" {
		t.Fatalf("unexpected left path: got %q want %q", result.Left, "left.env")
	}
	if result.Right != "right.env" {
		t.Fatalf("unexpected right path: got %q want %q", result.Right, "right.env")
	}
}

func assertResult(t *testing.T, result Result, wantMissingFromLeft, wantMissingFromRight []string) {
	t.Helper()

	if !slices.Equal(result.MissingFromLeft, wantMissingFromLeft) {
		t.Fatalf("unexpected MissingFromLeft: got %v want %v", result.MissingFromLeft, wantMissingFromLeft)
	}
	if !slices.Equal(result.MissingFromRight, wantMissingFromRight) {
		t.Fatalf("unexpected MissingFromRight: got %v want %v", result.MissingFromRight, wantMissingFromRight)
	}
}
