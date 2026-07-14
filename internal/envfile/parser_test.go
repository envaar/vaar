/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package envfile_test verifies that Parse keeps line metadata, quote state,
// BOMs, comments and newline shape intact.
package envfile_test

import (
	"testing"

	"github.com/envaar/vaar/internal/envfile"
)

func TestParseCapturesLineMetadata(t *testing.T) {
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("DATABASE_URL=postgres://localhost/db\r\nBAD: value\nNO_VALUE\n=value\nKEY_TWO=\"quoted\"\n\n")...)

	file, err := envfile.Parse("test.env", data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if !file.BOM {
		t.Fatalf("expected BOM to be recorded")
	}
	if !file.MixedLineEndings {
		t.Fatalf("expected mixed line endings to be recorded")
	}
	if !file.EndsWithNewline {
		t.Fatalf("expected final newline to be recorded")
	}
	if got, want := len(file.Lines), 6; got != want {
		t.Fatalf("unexpected line count: got %d want %d", got, want)
	}

	if got, want := file.Lines[0].Key, "DATABASE_URL"; got != want {
		t.Fatalf("unexpected key on line 1: got %q want %q", got, want)
	}
	if got, want := file.Lines[0].Value, "postgres://localhost/db"; got != want {
		t.Fatalf("unexpected value on line 1: got %q want %q", got, want)
	}
	if got, want := file.Lines[0].LineEnding, envfile.LineEndingCRLF; got != want {
		t.Fatalf("unexpected line ending on line 1: got %q want %q", got, want)
	}

	if got, want := file.Lines[1].DelimiterState, envfile.DelimiterColon; got != want {
		t.Fatalf("unexpected delimiter on line 2: got %q want %q", got, want)
	}
	if got, want := file.Lines[2].DelimiterState, envfile.DelimiterMissing; got != want {
		t.Fatalf("unexpected delimiter on line 3: got %q want %q", got, want)
	}
	if !file.Lines[3].HasValue || file.Lines[3].HasKey {
		t.Fatalf("expected line 4 to be a value without a key")
	}
	if got, want := file.Lines[4].QuoteState, envfile.QuoteDouble; got != want {
		t.Fatalf("unexpected quote state on line 5: got %q want %q", got, want)
	}
	if !file.Lines[5].IsBlank {
		t.Fatalf("expected final line to be blank")
	}
}

func TestParseCapturesInlineCommentsTrailingWhitespaceAndRepeatedBlankLines(t *testing.T) {
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("API_TOKEN=supersecret-token-123 # keep this masked\r\nTRAIL=value  \n\n\n")...)

	file, err := envfile.Parse("test.env", data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if !file.BOM {
		t.Fatalf("expected BOM to be recorded")
	}
	if !file.MixedLineEndings {
		t.Fatalf("expected mixed line endings to be recorded")
	}
	if !file.EndsWithNewline {
		t.Fatalf("expected final newline to be recorded")
	}
	if got, want := len(file.Lines), 4; got != want {
		t.Fatalf("unexpected line count: got %d want %d", got, want)
	}

	if got, want := file.Lines[0].CommentState, envfile.CommentInline; got != want {
		t.Fatalf("unexpected comment state on line 1: got %q want %q", got, want)
	}
	if got, want := file.Lines[0].Value, "supersecret-token-123"; got != want {
		t.Fatalf("unexpected value on line 1: got %q want %q", got, want)
	}
	if got, want := file.Lines[0].LineEnding, envfile.LineEndingCRLF; got != want {
		t.Fatalf("unexpected line ending on line 1: got %q want %q", got, want)
	}

	if got, want := file.Lines[1].TrailingWhitespace, "  "; got != want {
		t.Fatalf("unexpected trailing whitespace on line 2: got %q want %q", got, want)
	}
	if got, want := file.Lines[1].LineEnding, envfile.LineEndingLF; got != want {
		t.Fatalf("unexpected line ending on line 2: got %q want %q", got, want)
	}

	if !file.Lines[2].IsBlank || !file.Lines[3].IsBlank {
		t.Fatalf("expected repeated blank lines to be preserved: %#v", file.Lines[2:])
	}
}
