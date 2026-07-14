/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package envfile parses dotenv files into a line-aware model and preserves
// enough metadata to flag BOMs, delimiter mistakes, quote problems, repeated
// blanks and line-ending drift without guessing at shell semantics.
package envfile

// QuoteState records how Parse handled the value on a line.
type QuoteState string

const (
	// QuoteNone means Parse saw no wrapping quotes.
	QuoteNone QuoteState = "none"
	// QuoteSingle means Parse saw a balanced single-quoted value.
	QuoteSingle QuoteState = "single"
	// QuoteDouble means Parse saw a balanced double-quoted value.
	QuoteDouble QuoteState = "double"
	// QuoteUnbalanced means the value opened with a quote but never closed cleanly.
	QuoteUnbalanced QuoteState = "unbalanced"
)

// CommentState records whether Parse found a full-line or inline comment.
type CommentState string

const (
	// CommentNone means the line contains no comment.
	CommentNone CommentState = "none"
	// CommentFull means the line starts with a comment marker after indentation.
	CommentFull CommentState = "full"
	// CommentInline means the line contains a trailing comment after content.
	CommentInline CommentState = "inline"
)

// DelimiterState records the separator Parse detected on a line.
type DelimiterState string

const (
	// DelimiterNone means Parse found no delimiter.
	DelimiterNone DelimiterState = "none"
	// DelimiterEquals means Parse found an equals sign.
	DelimiterEquals DelimiterState = "equals"
	// DelimiterColon means Parse found a colon.
	DelimiterColon DelimiterState = "colon"
	// DelimiterMissing means the line looks like a key but has no delimiter.
	DelimiterMissing DelimiterState = "missing"
)

// LineEnding records the line ending Split found for one raw line.
type LineEnding string

const (
	// LineEndingNone means the file ended without a trailing newline.
	LineEndingNone LineEnding = "none"
	// LineEndingLF means the line ended with a Unix newline.
	LineEndingLF LineEnding = "lf"
	// LineEndingCRLF means the line ended with a Windows newline.
	LineEndingCRLF LineEnding = "crlf"
)

// Line captures one parsed line from a dotenv file along with the metadata
// needed for linting, reporting and safe rewrites.
type Line struct {
	Path                 string
	Number               int
	Raw                  string
	Content              string
	Key                  string
	KeyRaw               string
	Value                string
	ValueRaw             string
	LeadingWhitespace    string
	TrailingWhitespace   string
	QuoteState           QuoteState
	CommentState         CommentState
	DelimiterState       DelimiterState
	LineEnding           LineEnding
	HasKey               bool
	HasValue             bool
	IsBlank              bool
	IsComment            bool
	HasAssignment        bool
	SpaceBeforeDelimiter bool
	SpaceAfterDelimiter  bool
	BOM                  bool
}

// File captures one dotenv file, its parsed lines and whole-file metadata
// such as BOM presence, mixed endings and whether the file ends cleanly.
type File struct {
	Path             string
	BOM              bool
	MixedLineEndings bool
	EndsWithNewline  bool
	Lines            []Line
	Original         []byte
}
