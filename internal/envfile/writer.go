/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package envfile

import (
	"os"
	"strings"
)

// Normalize strips a UTF-8 BOM, trims trailing whitespace, collapses repeated
// blank lines and leaves exactly one trailing newline on non-empty files so
// --fix only makes safe, line-local edits.
func Normalize(data []byte) []byte {
	lines := Split(data)
	if len(lines) == 0 {
		return []byte{}
	}

	normalized := make([]string, 0, len(lines))
	prevBlank := false

	for i, raw := range lines {
		content := raw.Content
		if i == 0 && strings.HasPrefix(content, "\ufeff") {
			content = strings.TrimPrefix(content, "\ufeff")
		}

		content = strings.TrimRight(content, " \t")
		blank := strings.TrimSpace(content) == ""
		if blank {
			if prevBlank {
				continue
			}
			prevBlank = true
			normalized = append(normalized, "")
			continue
		}

		prevBlank = false
		normalized = append(normalized, content)
	}

	for len(normalized) > 0 && normalized[len(normalized)-1] == "" {
		normalized = normalized[:len(normalized)-1]
	}

	if len(normalized) == 0 {
		return []byte{}
	}

	var builder strings.Builder
	for _, line := range normalized {
		builder.WriteString(line)
		builder.WriteByte('\n')
	}

	return []byte(builder.String())
}

// Write writes data back to path and preserves the file's existing permissions
// when the file already exists.
func Write(path string, data []byte) error {
	perm := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}

	return os.WriteFile(path, data, perm)
}
