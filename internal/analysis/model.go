// Copyright © 2026 envaar
// SPDX-License-Identifier: Apache-2.0

package analysis

type DocumentID string

type QuoteState string

const (
	QuoteNone       QuoteState = "none"
	QuoteSingle     QuoteState = "single"
	QuoteDouble     QuoteState = "double"
	QuoteUnbalanced QuoteState = "unbalanced"
)

type CommentState string

const (
	CommentNone   CommentState = "none"
	CommentFull   CommentState = "full"
	CommentInline CommentState = "inline"
)

type DelimiterState string

const (
	DelimiterNone    DelimiterState = "none"
	DelimiterEquals  DelimiterState = "equals"
	DelimiterColon   DelimiterState = "colon"
	DelimiterMissing DelimiterState = "missing"
)

type LineEnding string

const (
	LineEndingNone LineEnding = "none"
	LineEndingLF   LineEnding = "lf"
	LineEndingCRLF LineEnding = "crlf"
)

type Line struct {
	Number                  int
	Key                     string
	HasKey                  bool
	HasValue                bool
	HasAssignment           bool
	IsBlank                 bool
	IsComment               bool
	QuoteState              QuoteState
	CommentState            CommentState
	DelimiterState          DelimiterState
	LineEnding              LineEnding
	HasLeadingWhitespace    bool
	HasTrailingWhitespace   bool
	SpaceBeforeDelimiter    bool
	SpaceAfterDelimiter     bool
	ValueContainsWhitespace bool
}

type Document struct {
	ID               DocumentID
	SourcePath       string
	DisplayPath      string
	BOM              bool
	MixedLineEndings bool
	EndsWithNewline  bool
	Lines            []Line
}

type Snapshot struct {
	documents []Document
}

func NewSnapshot(documents []Document) Snapshot {
	return Snapshot{documents: cloneDocuments(documents)}
}

func (s Snapshot) Documents() []Document {
	return cloneDocuments(s.documents)
}

func cloneDocuments(documents []Document) []Document {
	cloned := make([]Document, len(documents))
	for i, document := range documents {
		cloned[i] = document
		cloned[i].Lines = make([]Line, len(document.Lines))
		copy(cloned[i].Lines, document.Lines)
	}
	return cloned
}
