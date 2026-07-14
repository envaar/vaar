/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package envfile

// RawLine keeps one line of text and its detected line ending before semantic
// parsing starts.
type RawLine struct {
	Number  int
	Content string
	Ending  LineEnding
}

// Split breaks data into raw lines, preserves the line ending for each record,
// and keeps the final unterminated line instead of dropping it.
func Split(data []byte) []RawLine {
	if len(data) == 0 {
		return nil
	}

	lines := make([]RawLine, 0, 16)
	start := 0
	lineNumber := 1

	for start < len(data) {
		i := start
		for i < len(data) && data[i] != '\n' && data[i] != '\r' {
			i++
		}

		content := string(data[start:i])
		ending := LineEndingNone
		switch {
		case i < len(data) && data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n':
			ending = LineEndingCRLF
			i += 2
		case i < len(data) && data[i] == '\n':
			ending = LineEndingLF
			i++
		case i < len(data):
			i++
		}

		lines = append(lines, RawLine{
			Number:  lineNumber,
			Content: content,
			Ending:  ending,
		})

		lineNumber++
		start = i
	}

	return lines
}
