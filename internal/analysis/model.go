// Copyright © 2026 envaar
// SPDX-License-Identifier: Apache-2.0

package analysis

import (
	"sort"
	"strings"
)

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

// DocumentView exposes document metadata and read-only line traversal without
// exposing the snapshot's backing line slice.
type DocumentView struct {
	ID               DocumentID
	SourcePath       string
	DisplayPath      string
	BOM              bool
	MixedLineEndings bool
	EndsWithNewline  bool

	lines []Line
}

// RangeDocuments calls fn once for each document in source order. The view
// does not expose mutable collection storage, so callers can inspect a
// snapshot without allocating a complete defensive copy for every consumer.
func (s Snapshot) RangeDocuments(fn func(DocumentView)) {
	if fn == nil {
		return
	}

	for _, document := range s.documents {
		fn(DocumentView{
			ID:               document.ID,
			SourcePath:       document.SourcePath,
			DisplayPath:      document.DisplayPath,
			BOM:              document.BOM,
			MixedLineEndings: document.MixedLineEndings,
			EndsWithNewline:  document.EndsWithNewline,
			lines:            document.Lines,
		})
	}
}

// RangeLines calls fn once for each line in source order. Each line is passed
// by value, so changing the callback argument cannot mutate the snapshot.
func (d DocumentView) RangeLines(fn func(Line)) {
	if fn == nil {
		return
	}

	for _, line := range d.lines {
		fn(line)
	}
}

// Declaration identifies one value-free assignment occurrence in a document.
// It preserves only the document identity, user-facing path, and source line.
type Declaration struct {
	DocumentID  DocumentID
	DisplayPath string
	LineNumber  int
}

// KeyInventory is a read-only, document-local index of valid assignment keys
// and their declaration locations.
//
// A key is valid when its analysis line has both HasAssignment and HasKey set.
// This preserves the existing diff semantics: empty assignments count, while
// comments, blank lines, bare keys, and malformed declarations do not.
type KeyInventory struct {
	keys         []string
	declarations map[string][]Declaration
}

// NewKeyInventory builds a value-free inventory from one analysis document.
// The inventory retains no reference to the document's line slice. Keys are
// listed deterministically, while declarations retain source line order.
func NewKeyInventory(document Document) KeyInventory {
	declarations := make(map[string][]Declaration)

	for _, line := range document.Lines {
		if !line.HasAssignment || !line.HasKey {
			continue
		}

		key := strings.Clone(line.Key)
		declarations[key] = append(declarations[key], Declaration{
			DocumentID:  document.ID,
			DisplayPath: document.DisplayPath,
			LineNumber:  line.Number,
		})
	}

	keys := make([]string, 0, len(declarations))
	for key := range declarations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return KeyInventory{
		keys:         keys,
		declarations: declarations,
	}
}

// Contains reports whether key is present as a valid assignment key.
func (i KeyInventory) Contains(key string) bool {
	_, ok := i.declarations[key]
	return ok
}

// Keys returns all valid assignment keys in deterministic lexical order.
func (i KeyInventory) Keys() []string {
	keys := make([]string, len(i.keys))
	copy(keys, i.keys)
	return keys
}

// Declarations returns all source-order occurrences of key. The returned
// slice is independent of the inventory and is empty when key is absent.
func (i KeyInventory) Declarations(key string) []Declaration {
	declarations := make([]Declaration, len(i.declarations[key]))
	copy(declarations, i.declarations[key])
	return declarations
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
