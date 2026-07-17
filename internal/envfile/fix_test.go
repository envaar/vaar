/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package envfile_test verifies the per-rule fix transforms that decompose the
// historical whole-file Normalize pass.
package envfile_test

import (
	"testing"

	"github.com/envaar/vaar/internal/envfile"
)

// TestTrimTrailingWhitespacePreservesDelimiters pins that the transform strips
// spaces and tabs immediately before each line delimiter while preserving the
// delimiter itself, so a scoped --fix on a CRLF or lone-CR file repairs the
// trailing whitespace without rewriting the line endings. It also pins the
// whitespace-only-line emptying that the full-composition legacy parity relies
// on, including the vertical-tab and form-feed cases.
func TestTrimTrailingWhitespacePreservesDelimiters(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "crlf trailing spaces trimmed, crlf preserved",
			in:   "KEY=value  \r\nX=1\r\n",
			want: "KEY=value\r\nX=1\r\n",
		},
		{
			name: "crlf trailing tabs trimmed, crlf preserved",
			in:   "KEY=value\t\t\r\nX=1\r\n",
			want: "KEY=value\r\nX=1\r\n",
		},
		{
			name: "lone cr trailing spaces trimmed, cr preserved",
			in:   "KEY=value  \rX=1\r",
			want: "KEY=value\rX=1\r",
		},
		{
			name: "lf trailing spaces trimmed, lf preserved",
			in:   "KEY=value  \nX=1\n",
			want: "KEY=value\nX=1\n",
		},
		{
			name: "final unterminated line trimmed",
			in:   "KEY=value  ",
			want: "KEY=value",
		},
		{
			name: "clean crlf untouched",
			in:   "A=1\r\nB=2\r\n",
			want: "A=1\r\nB=2\r\n",
		},
		{
			name: "whitespace-only crlf line empties to its delimiter",
			in:   "A=1\r\n  \t\r\nB=2\r\n",
			want: "A=1\r\n\r\nB=2\r\n",
		},
		{
			name: "whitespace-only lf line empties to its delimiter",
			in:   "A=1\n  \t\nB=2\n",
			want: "A=1\n\nB=2\n",
		},
		{
			name: "vertical tab line empties, delimiter kept",
			in:   "A=1\n  \v  \nB=2\n",
			want: "A=1\n\nB=2\n",
		},
		{
			name: "form feed line empties, delimiter kept",
			in:   "A=1\n\f\nB=2\n",
			want: "A=1\n\nB=2\n",
		},
		{
			name: "empty input stays empty",
			in:   "",
			want: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := envfile.TrimTrailingWhitespace([]byte(tc.in))
			if string(got) != tc.want {
				t.Fatalf("TrimTrailingWhitespace(%q) = %q, want %q", tc.in, string(got), tc.want)
			}
		})
	}
}

// TestCollapseBlankLinesPreservesCRLF pins that the transform recognizes blank
// lines in a CRLF file (where splitting on LF leaves a CR on each segment) and
// collapses each run to one blank line, while retained lines keep their CR so
// the rejoined output preserves CRLF endings byte for byte.
func TestCollapseBlankLinesPreservesCRLF(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "consecutive blank crlf lines collapse to one",
			in:   "A=1\r\n\r\n\r\nB=2\r\n",
			want: "A=1\r\n\r\nB=2\r\n",
		},
		{
			name: "whitespace-only blank crlf lines collapse to one",
			in:   "A=1\r\n  \r\n\t\r\nB=2\r\n",
			want: "A=1\r\n  \r\nB=2\r\n",
		},
		{
			name: "single blank crlf line kept",
			in:   "A=1\r\n\r\nB=2\r\n",
			want: "A=1\r\n\r\nB=2\r\n",
		},
		{
			name: "crlf content lines byte exact",
			in:   "A=1\r\nB=2\r\n",
			want: "A=1\r\nB=2\r\n",
		},
		{
			name: "lf blank lines still collapse",
			in:   "A=1\n\n\nB=2\n",
			want: "A=1\n\nB=2\n",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := envfile.CollapseBlankLines([]byte(tc.in))
			if string(got) != tc.want {
				t.Fatalf("CollapseBlankLines(%q) = %q, want %q", tc.in, string(got), tc.want)
			}
		})
	}
}

// TestTrimFinalBlankLinesPreservesCRLF pins that the transform recognizes
// trailing blank lines in a CRLF file (where splitting on LF leaves a CR on
// each segment) and trims them to exactly one final newline, while a retained
// content line keeps its CR so its CRLF ending survives byte for byte.
func TestTrimFinalBlankLinesPreservesCRLF(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trailing blank crlf lines trimmed",
			in:   "A=1\r\n\r\n\r\n",
			want: "A=1\r\n",
		},
		{
			name: "trailing whitespace-only blank crlf lines trimmed",
			in:   "A=1\r\n  \r\n\t\r\n",
			want: "A=1\r\n",
		},
		{
			name: "missing final newline added, crlf kept",
			in:   "A=1\r\nB=2\r",
			want: "A=1\r\nB=2\r\n",
		},
		{
			name: "retained crlf content line byte exact",
			in:   "foo\r\n\r\n",
			want: "foo\r\n",
		},
		{
			name: "all-blank crlf file empties",
			in:   "\r\n  \r\n",
			want: "",
		},
		{
			name: "lf trailing blank lines still trimmed",
			in:   "A=1\n\n\n",
			want: "A=1\n",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := envfile.TrimFinalBlankLines([]byte(tc.in))
			if string(got) != tc.want {
				t.Fatalf("TrimFinalBlankLines(%q) = %q, want %q", tc.in, string(got), tc.want)
			}
		})
	}
}
