// Copyright © 2026 envaar
// SPDX-License-Identifier: Apache-2.0

package dotenv

import (
	"strings"

	"github.com/envaar/vaar/internal/analysis"
	sourcedotenv "github.com/envaar/vaar/internal/source/dotenv"
)

// DocumentInput pairs a caller-supplied stable identity with one loaded
// dotenv source document. The adapter never derives an identity from a path.
type DocumentInput struct {
	ID     analysis.DocumentID
	Source sourcedotenv.Document
}

// FromDocument converts one loaded dotenv source document into a value-free
// analysis document. It does not read files, validate parser output, or
// retain source values or other raw source data.
func FromDocument(input DocumentInput) analysis.Document {
	lines := make([]analysis.Line, len(input.Source.Lines))
	for i, sourceLine := range input.Source.Lines {
		lines[i] = analysis.Line{
			Number:                  sourceLine.Number,
			Key:                     strings.Clone(sourceLine.Key),
			HasKey:                  sourceLine.HasKey,
			HasValue:                sourceLine.HasValue,
			HasAssignment:           sourceLine.HasAssignment,
			IsBlank:                 sourceLine.IsBlank,
			IsComment:               sourceLine.IsComment,
			QuoteState:              analysis.QuoteState(sourceLine.QuoteState),
			CommentState:            analysis.CommentState(sourceLine.CommentState),
			DelimiterState:          analysis.DelimiterState(sourceLine.DelimiterState),
			LineEnding:              analysis.LineEnding(sourceLine.LineEnding),
			HasLeadingWhitespace:    sourceLine.LeadingWhitespace != "",
			HasTrailingWhitespace:   sourceLine.TrailingWhitespace != "",
			SpaceBeforeDelimiter:    sourceLine.SpaceBeforeDelimiter,
			SpaceAfterDelimiter:     sourceLine.SpaceAfterDelimiter,
			ValueContainsWhitespace: strings.ContainsAny(sourceLine.Value, " \t"),
		}
	}

	return analysis.Document{
		ID:               input.ID,
		SourcePath:       input.Source.SourcePath,
		DisplayPath:      input.Source.Path,
		BOM:              input.Source.BOM,
		MixedLineEndings: input.Source.MixedLineEndings,
		EndsWithNewline:  input.Source.EndsWithNewline,
		Lines:            lines,
	}
}

// FromDocuments converts loaded dotenv source documents in caller-supplied
// order. It returns a fresh non-nil empty slice for nil or empty input.
func FromDocuments(inputs []DocumentInput) []analysis.Document {
	documents := make([]analysis.Document, len(inputs))
	for i, input := range inputs {
		documents[i] = FromDocument(input)
	}
	return documents
}
