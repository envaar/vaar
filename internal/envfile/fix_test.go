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
			// A vertical tab or form feed is not a space or tab, so the
			// trailing-whitespace rule does not model it and this line is not a
			// whitespace-only finding. The fix keeps the line's non-space/tab
			// content (trailing spaces and tabs trimmed) instead of emptying it,
			// so it does not touch bytes the rule does not own.
			name: "vertical tab line kept, trailing spaces trimmed",
			in:   "A=1\n  \v  \nB=2\n",
			want: "A=1\n  \v\nB=2\n",
		},
		{
			name: "form feed line kept, delimiter unchanged",
			in:   "A=1\n\f\nB=2\n",
			want: "A=1\n\f\nB=2\n",
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

// TestCollapseBlankLinesPreservesLoneCR pins that the transform recognizes
// blank lines and line boundaries in a lone-CR file (CR as the only delimiter,
// no LF to split on) and collapses each run to one blank line, while retained
// lines keep their lone CR so the output preserves the lone-CR style byte for
// byte. Before the delimiter-complete walk, splitting on LF alone saw the whole
// file as one line and collapsed nothing, so a scoped --fix left the finding
// unresolved.
func TestCollapseBlankLinesPreservesLoneCR(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "consecutive blank cr lines collapse to one",
			in:   "KEY=1\r\r\rNEXT=2\r",
			want: "KEY=1\r\rNEXT=2\r",
		},
		{
			name: "whitespace-only blank cr lines collapse to one",
			in:   "A=1\r  \r\t\rB=2\r",
			want: "A=1\r  \rB=2\r",
		},
		{
			name: "single blank cr line kept",
			in:   "A=1\r\rB=2\r",
			want: "A=1\r\rB=2\r",
		},
		{
			name: "crlf blank lines still collapse",
			in:   "A=1\r\n\r\n\r\nB=2\r\n",
			want: "A=1\r\n\r\nB=2\r\n",
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

// TestTrimFinalBlankLinesPreservesLoneCR pins that the transform recognizes
// trailing blank lines in a lone-CR file (CR as the only delimiter, no LF to
// split on) and trims them, leaving exactly one lone-CR terminator on the last
// retained content line rather than injecting an LF. Before the delimiter-
// complete walk, splitting on LF alone saw the whole file as one line and
// trimmed nothing, so a scoped --fix left the finding unresolved.
func TestTrimFinalBlankLinesPreservesLoneCR(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			// A lone CR is itself a line terminator the lexer recognizes, so a
			// pure lone-CR file already ends with a newline. The terminal lone
			// CR is completed to CRLF only when the prevailing delimiter is CRLF
			// (see the crlf test below); with no prevailing CRLF it is kept as a
			// lone CR rather than being turned into A=1\r\n.
			name: "pure lone cr terminator preserved",
			in:   "A=1\r",
			want: "A=1\r",
		},
		{
			name: "trailing blank cr lines trimmed, lone cr preserved",
			in:   "A=1\r\r\r",
			want: "A=1\r",
		},
		{
			name: "trailing whitespace-only blank cr lines trimmed",
			in:   "A=1\r  \r\t\r",
			want: "A=1\r",
		},
		{
			name: "all-blank cr file empties",
			in:   "\r  \r\t\r",
			want: "",
		},
		{
			name: "crlf trailing blank lines still trimmed",
			in:   "A=1\r\n\r\n\r\n",
			want: "A=1\r\n",
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
			name: "unterminated final line adopts prevailing crlf",
			in:   "A=1\r\nB=2",
			want: "A=1\r\nB=2\r\n",
		},
		{
			name: "unterminated final line adopts prevailing lf",
			in:   "A=1\nB=2",
			want: "A=1\nB=2\n",
		},
		{
			name: "single unterminated line defaults to lf",
			in:   "A=1",
			want: "A=1\n",
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

// TestHasMixedLineEndings pins the condition the finding-scoped line-ending fix
// uses: it must match the parser's File.MixedLineEndings (sawLF && sawCRLF)
// exactly, so a file mixes endings only when it carries both an LF-terminated
// and a CRLF-terminated line. A file with uniform endings — all LF, all CRLF, or
// all lone CR — is not mixed and so has no line-ending finding to fix.
func TestHasMixedLineEndings(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"uniform lf", "A=1\nB=2\n", false},
		{"uniform crlf", "A=1\r\nB=2\r\n", false},
		{"uniform lone cr", "A=1\rB=2\r", false},
		{"crlf and lf mixed", "A=1\r\nB=2\nC=3\r\n", true},
		{"lf then crlf mixed", "A=1\nB=2\r\n", true},
		{"lone cr with lf not mixed", "A=1\rB=2\n", false},
		{"lone cr with crlf not mixed", "A=1\rB=2\r\n", false},
		{"empty not mixed", "", false},
		{"single unterminated line not mixed", "A=1", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := envfile.HasMixedLineEndings([]byte(tc.in)); got != tc.want {
				t.Fatalf("HasMixedLineEndings(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
