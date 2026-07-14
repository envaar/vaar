/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package envfile

import "strings"

// Parse converts dotenv bytes into a semantic file model that keeps the
// original bytes, the parsed lines and the metadata needed for lint rules.
func Parse(path string, data []byte) (File, error) {
	rawLines := Split(data)
	file := File{
		Path:     path,
		Original: append([]byte(nil), data...),
		Lines:    make([]Line, 0, len(rawLines)),
	}

	if len(rawLines) == 0 {
		return file, nil
	}

	var sawLF bool
	var sawCRLF bool

	for i, raw := range rawLines {
		line := parseLine(path, raw, i == 0)
		if line.BOM {
			file.BOM = true
		}
		switch line.LineEnding {
		case LineEndingLF:
			sawLF = true
		case LineEndingCRLF:
			sawCRLF = true
		}
		file.Lines = append(file.Lines, line)
	}

	file.MixedLineEndings = sawLF && sawCRLF
	file.EndsWithNewline = rawLines[len(rawLines)-1].Ending != LineEndingNone

	return file, nil
}

// parseLine turns one raw line into the semantic fields used by lint rules and
// report rendering.
func parseLine(path string, raw RawLine, first bool) Line {
	line := Line{
		Path:       path,
		Number:     raw.Number,
		Raw:        raw.Content,
		Content:    raw.Content,
		LineEnding: raw.Ending,
	}

	content := raw.Content
	if first && strings.HasPrefix(content, "\ufeff") {
		// Only the first line can carry a BOM and we strip it from the parsed
		// content while leaving the original bytes untouched in File.Original.
		line.BOM = true
		content = strings.TrimPrefix(content, "\ufeff")
		line.Content = content
	}

	line.LeadingWhitespace = leadingWhitespace(content)
	line.TrailingWhitespace = trailingWhitespace(content)

	trimmedLeft := strings.TrimLeft(content, " \t")
	if trimmedLeft == "" {
		line.IsBlank = true
		return line
	}

	if strings.HasPrefix(trimmedLeft, "#") {
		// Treat indentation followed by # as a full-line comment.
		line.IsComment = true
		line.CommentState = CommentFull
		return line
	}

	body := trimmedLeft
	delimIdx, delim, commentIdx := scanLine(body)
	if commentIdx >= 0 {
		// Strip the inline comment from the semantic body but keep the raw text
		// on the line for diagnostics and formatting checks.
		body = body[:commentIdx]
	}

	if delimIdx < 0 {
		key := strings.TrimSpace(body)
		if key != "" && isKeyLike(key) {
			// Key-like lines without a delimiter usually signal a missing value.
			line.Key = key
			line.KeyRaw = body
			line.HasKey = true
			line.DelimiterState = DelimiterMissing
			return line
		}

		line.Value = strings.TrimSpace(body)
		line.ValueRaw = body
		line.HasValue = line.Value != ""
		return line
	}

	keyRaw := body[:delimIdx]
	line.KeyRaw = keyRaw
	line.Key = strings.TrimSpace(keyRaw)
	line.HasKey = line.Key != ""
	line.SpaceBeforeDelimiter = strings.TrimRight(keyRaw, " \t") != keyRaw
	line.DelimiterState = delimiterState(delim)
	line.HasAssignment = delim == '='

	valueRaw := body[delimIdx+1:]
	line.ValueRaw = valueRaw
	line.SpaceAfterDelimiter = strings.HasPrefix(valueRaw, " ") || strings.HasPrefix(valueRaw, "\t")

	value := strings.TrimSpace(valueRaw)
	if value == "" {
		return line
	}

	line.HasValue = true
	quoteState, parsedValue := parseQuotedValue(value)
	line.QuoteState = quoteState
	line.Value = parsedValue
	if quoteState == QuoteNone {
		line.Value = value
	}

	if commentIdx >= 0 {
		line.CommentState = CommentInline
	}

	return line
}

func leadingWhitespace(s string) string {
	i := 0
	for i < len(s) {
		if s[i] != ' ' && s[i] != '\t' {
			break
		}
		i++
	}
	return s[:i]
}

func trailingWhitespace(s string) string {
	i := len(s)
	for i > 0 {
		if s[i-1] != ' ' && s[i-1] != '\t' {
			break
		}
		i--
	}
	return s[i:]
}

// scanLine finds the first delimiter and the first comment marker outside of
// quoted text.
func scanLine(s string) (delimIdx int, delim rune, commentIdx int) {
	delimIdx = -1
	commentIdx = -1

	inSingle := false
	inDouble := false

	for i, r := range s {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				// Stop at the first comment marker that appears outside quotes.
				commentIdx = i
				return
			}
		case '=', ':':
			if !inSingle && !inDouble && delimIdx == -1 {
				// Keep the first delimiter outside quotes. The parser splits on the
				// earliest assignment marker it sees.
				delimIdx = i
				delim = r
			}
		}
	}

	return
}

// parseQuotedValue unwraps balanced quotes and flags malformed tails or
// unmatched openings.
func parseQuotedValue(value string) (QuoteState, string) {
	if value == "" {
		return QuoteNone, ""
	}

	first := value[0]
	if first != '"' && first != '\'' {
		return QuoteNone, value
	}

	end := -1
	escaped := false
	for i := 1; i < len(value); i++ {
		ch := value[i]
		if first == '"' && ch == '\\' && !escaped {
			// Double quotes honor backslash escapes for embedded quotes.
			escaped = true
			continue
		}
		if ch == first && !escaped {
			end = i
			break
		}
		escaped = false
	}

	if end == -1 {
		// Keep the inner text so rules can report the malformed value without
		// losing the original content.
		return QuoteUnbalanced, value[1:]
	}

	if tail := strings.TrimSpace(value[end+1:]); tail != "" {
		// Any extra content after the closing quote makes the value malformed.
		return QuoteUnbalanced, value[1:end]
	}

	inner := value[1:end]
	if first == '"' {
		inner = strings.ReplaceAll(inner, `\"`, `"`)
	}

	switch first {
	case '"':
		return QuoteDouble, inner
	default:
		return QuoteSingle, inner
	}
}

// delimiterState converts the detected delimiter rune into the exported state
// used by rules and reports.
func delimiterState(delim rune) DelimiterState {
	switch delim {
	case '=':
		return DelimiterEquals
	case ':':
		return DelimiterColon
	default:
		return DelimiterNone
	}
}

// isKeyLike accepts the key shape that looks like a dotenv assignment target.
func isKeyLike(s string) bool {
	if s == "" {
		return false
	}

	for i, r := range s {
		switch {
		case i == 0:
			if !isKeyStart(r) {
				return false
			}
		default:
			if !isKeyChar(r) {
				return false
			}
		}
	}

	return true
}

func isKeyStart(r rune) bool {
	return r == '_' || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z')
}

func isKeyChar(r rune) bool {
	return isKeyStart(r) || ('0' <= r && r <= '9')
}
