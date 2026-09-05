// Copyright © 2026 envaar
// SPDX-License-Identifier: Apache-2.0

package analysis_test

import (
	"reflect"
	"testing"

	"github.com/envaar/vaar/internal/analysis"
)

func TestNewSnapshotEmpty(t *testing.T) {
	snapshot := analysis.NewSnapshot(nil)

	documents := snapshot.Documents()
	if documents == nil {
		t.Fatal("Documents() returned nil; want non-nil empty slice")
	}
	if len(documents) != 0 {
		t.Fatalf("Documents() returned %d documents; want 0", len(documents))
	}
}

func TestSnapshotPreservesEmptyDocument(t *testing.T) {
	document := analysis.Document{
		ID:               analysis.DocumentID("doc-empty"),
		SourcePath:       "/work/.env",
		DisplayPath:      ".env",
		BOM:              true,
		MixedLineEndings: true,
		EndsWithNewline:  true,
		Lines:            []analysis.Line{},
	}

	snapshot := analysis.NewSnapshot([]analysis.Document{document})
	documents := snapshot.Documents()

	if len(documents) != 1 {
		t.Fatalf("Documents() returned %d documents; want 1", len(documents))
	}
	if !reflect.DeepEqual(documents[0], document) {
		t.Fatalf("document mismatch\n got: %#v\nwant: %#v", documents[0], document)
	}
	if documents[0].Lines == nil {
		t.Fatal("empty document lines returned nil; want non-nil empty slice")
	}
}

func TestSnapshotPreservesDocumentOrderAndIdentity(t *testing.T) {
	documents := []analysis.Document{
		{
			ID:          analysis.DocumentID("zeta"),
			SourcePath:  "/repo/z.env",
			DisplayPath: "z.env",
		},
		{
			ID:          analysis.DocumentID("alpha"),
			SourcePath:  "/repo/a.env",
			DisplayPath: "a.env",
		},
	}

	got := analysis.NewSnapshot(documents).Documents()

	if len(got) != len(documents) {
		t.Fatalf("Documents() returned %d documents; want %d", len(got), len(documents))
	}
	for i := range documents {
		if got[i].ID != documents[i].ID {
			t.Fatalf("document %d ID = %q; want %q", i, got[i].ID, documents[i].ID)
		}
		if got[i].SourcePath != documents[i].SourcePath {
			t.Fatalf("document %d SourcePath = %q; want %q", i, got[i].SourcePath, documents[i].SourcePath)
		}
		if got[i].DisplayPath != documents[i].DisplayPath {
			t.Fatalf("document %d DisplayPath = %q; want %q", i, got[i].DisplayPath, documents[i].DisplayPath)
		}
	}
}

func TestSnapshotPreservesLineOrderAndProvenance(t *testing.T) {
	document := analysis.Document{
		ID:          analysis.DocumentID("doc-lines"),
		SourcePath:  "/repo/service/.env",
		DisplayPath: "service/.env",
		Lines: []analysis.Line{
			{Number: 1, Key: "FIRST", HasKey: true},
			{Number: 2, Key: "SECOND", HasKey: true},
			{Number: 7, Key: "SEVENTH", HasKey: true},
		},
	}

	documents := analysis.NewSnapshot([]analysis.Document{document}).Documents()

	if len(documents) != 1 {
		t.Fatalf("Documents() returned %d documents; want 1", len(documents))
	}
	if documents[0].SourcePath != document.SourcePath {
		t.Fatalf("SourcePath = %q; want %q", documents[0].SourcePath, document.SourcePath)
	}
	if documents[0].DisplayPath != document.DisplayPath {
		t.Fatalf("DisplayPath = %q; want %q", documents[0].DisplayPath, document.DisplayPath)
	}
	if len(documents[0].Lines) != len(document.Lines) {
		t.Fatalf("line count = %d; want %d", len(documents[0].Lines), len(document.Lines))
	}
	for i := range document.Lines {
		if documents[0].Lines[i].Number != document.Lines[i].Number {
			t.Fatalf("line %d number = %d; want %d", i, documents[0].Lines[i].Number, document.Lines[i].Number)
		}
		if documents[0].Lines[i].Key != document.Lines[i].Key {
			t.Fatalf("line %d key = %q; want %q", i, documents[0].Lines[i].Key, document.Lines[i].Key)
		}
	}
}

func TestSnapshotPreservesSafeSyntaxFacts(t *testing.T) {
	lines := []analysis.Line{
		{
			Number:              1,
			Key:                 "POPULATED_EQUALS",
			HasKey:              true,
			HasValue:            true,
			HasAssignment:       true,
			QuoteState:          analysis.QuoteNone,
			CommentState:        analysis.CommentNone,
			DelimiterState:      analysis.DelimiterEquals,
			LineEnding:          analysis.LineEndingLF,
			SpaceAfterDelimiter: true,
		},
		{
			Number:               2,
			Key:                  "EMPTY_EQUALS",
			HasKey:               true,
			HasValue:             false,
			HasAssignment:        true,
			QuoteState:           analysis.QuoteNone,
			CommentState:         analysis.CommentNone,
			DelimiterState:       analysis.DelimiterEquals,
			LineEnding:           analysis.LineEndingCRLF,
			SpaceBeforeDelimiter: true,
		},
		{
			Number:         3,
			Key:            "BARE_KEY",
			HasKey:         true,
			HasValue:       false,
			HasAssignment:  false,
			QuoteState:     analysis.QuoteNone,
			CommentState:   analysis.CommentNone,
			DelimiterState: analysis.DelimiterMissing,
			LineEnding:     analysis.LineEndingNone,
		},
		{
			Number:               4,
			IsComment:            true,
			QuoteState:           analysis.QuoteNone,
			CommentState:         analysis.CommentFull,
			DelimiterState:       analysis.DelimiterNone,
			LineEnding:           analysis.LineEndingLF,
			HasLeadingWhitespace: true,
		},
		{
			Number:         5,
			IsBlank:        true,
			QuoteState:     analysis.QuoteNone,
			CommentState:   analysis.CommentNone,
			DelimiterState: analysis.DelimiterNone,
			LineEnding:     analysis.LineEndingLF,
		},
		{
			Number:              6,
			Key:                 "COLON_DELIMITER",
			HasKey:              true,
			HasValue:            true,
			HasAssignment:       true,
			QuoteState:          analysis.QuoteSingle,
			CommentState:        analysis.CommentInline,
			DelimiterState:      analysis.DelimiterColon,
			LineEnding:          analysis.LineEndingLF,
			SpaceAfterDelimiter: true,
		},
		{
			Number:                7,
			Key:                   "UNBALANCED_QUOTE",
			HasKey:                true,
			HasValue:              true,
			HasAssignment:         true,
			QuoteState:            analysis.QuoteUnbalanced,
			CommentState:          analysis.CommentNone,
			DelimiterState:        analysis.DelimiterEquals,
			LineEnding:            analysis.LineEndingLF,
			HasTrailingWhitespace: true,
		},
		{
			Number:                  8,
			Key:                     "VALUE_WHITESPACE",
			HasKey:                  true,
			HasValue:                true,
			HasAssignment:           true,
			QuoteState:              analysis.QuoteDouble,
			CommentState:            analysis.CommentNone,
			DelimiterState:          analysis.DelimiterEquals,
			LineEnding:              analysis.LineEndingLF,
			ValueContainsWhitespace: true,
		},
	}
	document := analysis.Document{
		ID:    analysis.DocumentID("syntax"),
		Lines: lines,
	}

	got := analysis.NewSnapshot([]analysis.Document{document}).Documents()

	if len(got) != 1 {
		t.Fatalf("Documents() returned %d documents; want 1", len(got))
	}
	if !reflect.DeepEqual(got[0].Lines, lines) {
		t.Fatalf("syntax facts mismatch\n got: %#v\nwant: %#v", got[0].Lines, lines)
	}
}

func TestSnapshotDoesNotExposeForbiddenValueFields(t *testing.T) {
	forbidden := []string{
		"Raw",
		"Content",
		"Value",
		"ValueRaw",
		"KeyRaw",
		"Original",
		"ValuePreview",
		"ValueLength",
		"Hash",
		"Mask",
	}

	assertFieldsAbsent(t, reflect.TypeOf(analysis.Document{}), forbidden)
	assertFieldsAbsent(t, reflect.TypeOf(analysis.Line{}), forbidden)
	if _, ok := reflect.TypeOf(analysis.Line{}).FieldByName("ValueContainsWhitespace"); !ok {
		t.Fatal("Line.ValueContainsWhitespace is missing")
	}
}

func TestSnapshotDefensivelyCopiesInput(t *testing.T) {
	input := []analysis.Document{
		{
			ID:          analysis.DocumentID("original"),
			SourcePath:  "/repo/original.env",
			DisplayPath: "original.env",
			Lines: []analysis.Line{
				{Number: 1, Key: "ORIGINAL", HasKey: true},
				{Number: 2, Key: "KEEP", HasKey: true},
			},
		},
	}
	snapshot := analysis.NewSnapshot(input)

	input[0].ID = analysis.DocumentID("mutated")
	input[0].SourcePath = "/repo/mutated.env"
	input[0].DisplayPath = "mutated.env"
	input[0].Lines[0].Number = 99
	input[0].Lines[0].Key = "MUTATED"
	input[0].Lines = input[0].Lines[:1]

	got := snapshot.Documents()
	if got[0].ID != analysis.DocumentID("original") {
		t.Fatalf("ID = %q; want original", got[0].ID)
	}
	if got[0].SourcePath != "/repo/original.env" {
		t.Fatalf("SourcePath = %q; want /repo/original.env", got[0].SourcePath)
	}
	if got[0].DisplayPath != "original.env" {
		t.Fatalf("DisplayPath = %q; want original.env", got[0].DisplayPath)
	}
	if len(got[0].Lines) != 2 {
		t.Fatalf("line count = %d; want 2", len(got[0].Lines))
	}
	if got[0].Lines[0].Number != 1 {
		t.Fatalf("first line number = %d; want 1", got[0].Lines[0].Number)
	}
	if got[0].Lines[0].Key != "ORIGINAL" {
		t.Fatalf("first line key = %q; want ORIGINAL", got[0].Lines[0].Key)
	}
}

func TestDocumentsReturnsDefensiveCopies(t *testing.T) {
	snapshot := analysis.NewSnapshot([]analysis.Document{
		{
			ID:          analysis.DocumentID("original"),
			SourcePath:  "/repo/original.env",
			DisplayPath: "original.env",
			Lines: []analysis.Line{
				{Number: 1, Key: "ORIGINAL", HasKey: true},
				{Number: 2, Key: "KEEP", HasKey: true},
			},
		},
	})

	first := snapshot.Documents()
	first[0].ID = analysis.DocumentID("mutated")
	first[0].SourcePath = "/repo/mutated.env"
	first[0].DisplayPath = "mutated.env"
	first[0].Lines[0].Number = 99
	first[0].Lines[0].Key = "MUTATED"
	first[0].Lines = first[0].Lines[:1]

	second := snapshot.Documents()
	if second[0].ID != analysis.DocumentID("original") {
		t.Fatalf("ID = %q; want original", second[0].ID)
	}
	if second[0].SourcePath != "/repo/original.env" {
		t.Fatalf("SourcePath = %q; want /repo/original.env", second[0].SourcePath)
	}
	if second[0].DisplayPath != "original.env" {
		t.Fatalf("DisplayPath = %q; want original.env", second[0].DisplayPath)
	}
	if len(second[0].Lines) != 2 {
		t.Fatalf("line count = %d; want 2", len(second[0].Lines))
	}
	if second[0].Lines[0].Number != 1 {
		t.Fatalf("first line number = %d; want 1", second[0].Lines[0].Number)
	}
	if second[0].Lines[0].Key != "ORIGINAL" {
		t.Fatalf("first line key = %q; want ORIGINAL", second[0].Lines[0].Key)
	}
}

func TestSnapshotHasNoExportedCollectionFields(t *testing.T) {
	snapshotType := reflect.TypeOf(analysis.Snapshot{})

	for i := 0; i < snapshotType.NumField(); i++ {
		field := snapshotType.Field(i)
		if field.IsExported() {
			t.Fatalf("Snapshot exposes exported field %q; want no exported fields", field.Name)
		}
	}
}

func assertFieldsAbsent(t *testing.T, typ reflect.Type, names []string) {
	t.Helper()

	for _, name := range names {
		if _, ok := typ.FieldByName(name); ok {
			t.Fatalf("%s exposes forbidden field %q", typ.Name(), name)
		}
	}
}
