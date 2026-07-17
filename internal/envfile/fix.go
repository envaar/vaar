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

// TrimTrailingWhitespace removes trailing spaces and tabs from every line and
// empties any line that is nothing but whitespace. It is the fix half of the
// trailing-whitespace rule.
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

		if len(bytes.TrimSpace(content)) != 0 {
			// A line with real content keeps it with trailing spaces and tabs
			// stripped; a whitespace-only line (including stray vertical tabs
			// or form feeds) carries none, so only its delimiter survives.
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
// Splitting on LF leaves a CR at the end of each line of a CRLF file, so the
// blank check strips one trailing CR first; the stored line keeps its CR, so
// retained lines rejoin with their CRLF endings intact.
func CollapseBlankLines(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	out := make([][]byte, 0, len(lines))
	prevBlank := false
	for _, line := range lines {
		if isBlankLine(bytes.TrimSuffix(line, []byte("\r"))) {
			if prevBlank {
				continue
			}
			prevBlank = true
		} else {
			prevBlank = false
		}
		out = append(out, line)
	}
	return bytes.Join(out, []byte("\n"))
}

// TrimFinalBlankLines drops trailing blank lines and leaves exactly one final
// newline on non-empty content, returning empty bytes for an all-blank or empty
// file. It is the fix half of the ending-blank-line rule.
//
// Splitting on LF leaves a CR at the end of each line of a CRLF file, so the
// blank check strips one trailing CR first; the stored line keeps its CR, so a
// retained final line is written back with its CRLF ending intact.
func TrimFinalBlankLines(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	end := len(lines)
	for end > 0 && isBlankLine(bytes.TrimSuffix(lines[end-1], []byte("\r"))) {
		end--
	}
	if end == 0 {
		return []byte{}
	}

	var buf bytes.Buffer
	for _, line := range lines[:end] {
		buf.Write(line)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// isBlankLine reports whether a raw line holds only spaces and tabs, matching
// the parser's IsBlank so the fix transforms and the rules agree on blankness.
func isBlankLine(line []byte) bool {
	return len(bytes.TrimLeft(line, " \t")) == 0
}
