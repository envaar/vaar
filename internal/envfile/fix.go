/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package envfile

import "bytes"

// The functions in this file decompose the historical whole-file Normalize
// pass into one surgical transform per fixable lint rule. Each transform only
// repairs the single defect its rule reports and leaves every other byte
// untouched, so callers can compose a subset to scope --fix to selected rules.
//
// Composing all of them in the canonical pipeline order
//
//	StripBOM -> NormalizeLineEndings -> TrimTrailingWhitespace ->
//	CollapseBlankLines -> TrimFinalBlankLines
//
// reproduces the old Normalize output byte for byte. The equivalence is pinned
// by TestFixDataMatchesLegacyNormalize in internal/lint/rules against a frozen
// copy of the original Normalize, so this decomposition cannot silently drift.

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// StripBOM removes a leading UTF-8 BOM so the first key on disk is not silently
// altered by the marker. It is the fix half of the bom-character rule.
func StripBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, utf8BOM)
}

// NormalizeLineEndings converts CRLF and lone CR line endings to LF so a file
// does not switch newline style midstream. It is the fix half of the
// line-ending rule.
func NormalizeLineEndings(data []byte) []byte {
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	return data
}

// HasMixedLineEndings reports whether data mixes LF and CRLF line endings. It
// mirrors exactly how the parser sets File.MixedLineEndings — sawLF && sawCRLF
// over the split lines — so a scoped line-ending fix can stay finding-conditional
// and agree byte for byte with what the rule reports. The line-ending rule flags
// a file only when it is mixed, so its fix normalizes only mixed input; a file
// whose endings are uniform (all LF, all CRLF, or all lone CR) has no finding and
// must be left untouched rather than blindly rewritten to LF.
func HasMixedLineEndings(data []byte) bool {
	var sawLF, sawCRLF bool
	for _, line := range Split(data) {
		switch line.Ending {
		case LineEndingLF:
			sawLF = true
		case LineEndingCRLF:
			sawCRLF = true
		}
	}
	return sawLF && sawCRLF
}

// TrimTrailingWhitespace removes trailing spaces and tabs from every line and
// empties any line that is nothing but spaces and tabs. It is the fix half of
// the trailing-whitespace rule.
//
// It splits on each line delimiter (LF, CRLF or a lone CR) and preserves that
// delimiter verbatim, stripping only the spaces and tabs immediately before it,
// so "KEY=value  \r\n" becomes "KEY=value\r\n". Trimming before the delimiter
// (rather than splitting on LF alone) lets a scoped --fix repair a CRLF file
// whose trailing whitespace sits ahead of a surviving CR. The full fix pipeline
// runs NormalizeLineEndings first, so no CR reaches this transform there and
// the composed output is unchanged; the CR handling matters only when this
// rule's fix runs on its own.
func TrimTrailingWhitespace(data []byte) []byte {
	var out bytes.Buffer
	for i := 0; i < len(data); {
		start := i
		for i < len(data) && data[i] != '\n' && data[i] != '\r' {
			i++
		}
		content := data[start:i]

		var delim []byte
		if i < len(data) {
			if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
				delim = data[i : i+2]
				i += 2
			} else {
				delim = data[i : i+1]
				i++
			}
		}

		if !isBlankLine(content) {
			// A line with any non-space/tab byte keeps its content with only
			// trailing spaces and tabs stripped; a line that is nothing but
			// spaces and tabs empties so only its delimiter survives. Blankness
			// uses the spaces/tabs-only isBlankLine definition the trailing-
			// whitespace rule models, so a line whose only "blank" bytes are a
			// vertical tab or form feed — which the rule does not own — is kept
			// (with its trailing spaces and tabs trimmed) rather than emptied.
			out.Write(bytes.TrimRight(content, " \t"))
		}
		out.Write(delim)
	}
	return out.Bytes()
}

// CollapseBlankLines collapses each run of consecutive blank lines to a single
// blank line. It is the fix half of the extra-blank-line rule. A line counts as
// blank when it holds only spaces and tabs, matching the parser's IsBlank.
//
// It walks each line delimiter (LF, CRLF or a lone CR) the same way the lexer
// does, so blank runs and line boundaries are recognized for every ending
// style, and it preserves each retained line's delimiter bytes verbatim: a CRLF
// file stays CRLF and a lone-CR file stays lone-CR. Splitting on LF alone would
// see a lone-CR file as one line and collapse nothing; the full fix pipeline
// runs NormalizeLineEndings first, so no CR reaches this transform there and the
// composed output is unchanged. The CR handling matters only when this rule's
// fix runs on its own under a scoped --fix.
func CollapseBlankLines(data []byte) []byte {
	var out bytes.Buffer
	prevBlank := false
	for i := 0; i < len(data); {
		content, delim, next := nextRawLine(data, i)
		i = next

		if isBlankLine(content) {
			if prevBlank {
				// Drop the duplicate blank line entirely, delimiter included.
				continue
			}
			prevBlank = true
		} else {
			prevBlank = false
		}
		out.Write(content)
		out.Write(delim)
	}
	return out.Bytes()
}

// TrimFinalBlankLines drops trailing blank lines and leaves exactly one final
// line terminator on non-empty content, returning empty bytes for an all-blank
// or empty file. It is the fix half of the ending-blank-line rule.
//
// It walks each line delimiter (LF, CRLF or a lone CR) the same way the lexer
// does, so trailing blank lines are recognized for every ending style, and the
// retained final terminator preserves the delimiter of the last kept content
// line rather than hard-coding LF: a CRLF file keeps its CRLF and a lone-CR file
// keeps its lone CR. A last content line with no delimiter of its own adopts the
// file's prevailing line ending (the last delimited line's delimiter) so a
// scoped fix on a CRLF file that merely lacks its final newline stays all-CRLF
// instead of introducing a bare LF; with no delimiter anywhere it defaults to a
// single LF. A lone CR that is the final byte of the input is itself a line
// terminator the lexer recognizes, so a pure lone-CR file already ends with a
// newline and is kept as-is; the terminal lone CR is completed to CRLF only when
// the prevailing delimiter is CRLF, i.e. it is the dangling half of the file's
// CRLF style. The full fix pipeline runs NormalizeLineEndings first, so no CR
// reaches this transform there and the composed output is unchanged; the CR
// handling matters only under a scoped --fix.
func TrimFinalBlankLines(data []byte) []byte {
	type rawLine struct{ content, delim []byte }

	var lines []rawLine
	for i := 0; i < len(data); {
		content, delim, next := nextRawLine(data, i)
		i = next
		lines = append(lines, rawLine{content: content, delim: delim})
	}

	end := len(lines)
	for end > 0 && isBlankLine(lines[end-1].content) {
		end--
	}
	if end == 0 {
		return []byte{}
	}

	var buf bytes.Buffer
	for j := 0; j < end; j++ {
		buf.Write(lines[j].content)
		if j < end-1 {
			buf.Write(lines[j].delim)
			continue
		}

		// Last retained content line: leave exactly one terminator.
		delim := lines[j].delim
		switch {
		case len(delim) == 0:
			// An unterminated final line adopts the file's prevailing line
			// ending — the delimiter of the last delimited line — so a scoped
			// fix on a CRLF file missing only its final newline stays all-CRLF
			// rather than introducing a bare LF. With no delimiter anywhere
			// there is no prevailing style, so default to a single LF.
			terminator := []byte("\n")
			for k := end - 2; k >= 0; k-- {
				if len(lines[k].delim) > 0 {
					terminator = lines[k].delim
					break
				}
			}
			buf.Write(terminator)
		case bytes.Equal(delim, []byte("\r")) && j == len(lines)-1:
			// A lone CR that is the final byte of the input is itself a line
			// terminator the lexer recognizes, so a pure lone-CR file already
			// ends with a newline and keeps its lone CR. Complete it to CRLF
			// only when the prevailing delimiter is CRLF — the terminal CR is
			// then the dangling half of the file's CRLF style. A lone CR that
			// instead separated this line from a dropped blank run keeps its
			// lone CR (below).
			prevailingCRLF := false
			for k := end - 2; k >= 0; k-- {
				if len(lines[k].delim) > 0 {
					prevailingCRLF = bytes.Equal(lines[k].delim, []byte("\r\n"))
					break
				}
			}
			if prevailingCRLF {
				buf.WriteByte('\r')
				buf.WriteByte('\n')
			} else {
				buf.WriteByte('\r')
			}
		default:
			buf.Write(delim)
		}
	}
	return buf.Bytes()
}

// nextRawLine returns the content and delimiter of the line starting at start,
// plus the index of the next line. content runs up to the next LF or CR and
// excludes the delimiter; delim is the CRLF, lone CR or LF that ends the line,
// or empty when the line is the final unterminated one. This mirrors the lexer's
// line model and TrimTrailingWhitespace's delimiter-preserving walk so the blank
// fixes stay in lock-step with what the rules report.
func nextRawLine(data []byte, start int) (content, delim []byte, next int) {
	i := start
	for i < len(data) && data[i] != '\n' && data[i] != '\r' {
		i++
	}
	content = data[start:i]

	if i < len(data) {
		if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
			delim = data[i : i+2]
			i += 2
		} else {
			delim = data[i : i+1]
			i++
		}
	}
	return content, delim, i
}

// isBlankLine reports whether a raw line holds only spaces and tabs, matching
// the parser's IsBlank so the fix transforms and the rules agree on blankness.
func isBlankLine(line []byte) bool {
	return len(bytes.TrimLeft(line, " \t")) == 0
}
