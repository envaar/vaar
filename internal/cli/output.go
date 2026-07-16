/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// validateOutputDestination rejects --output values that name a directory: a
// trailing path separator or a path that already exists as a directory. Without
// this guard the write path below creates a temporary file, fails to rename it
// onto the directory and then removes the directory, destroying it.
func validateOutputDestination(output string) error {
	if hasTrailingSeparator(output) {
		return NewToolError(fmt.Sprintf("cannot write lint output to %q: the path is a directory", output), nil)
	}
	if info, err := os.Stat(output); err == nil && info.IsDir() {
		return NewToolError(fmt.Sprintf("cannot write lint output to %q: the path is a directory", output), nil)
	}
	return nil
}

// writeJSONOutput writes data to path via a temporary file in the destination
// directory followed by an atomic rename. filepath.Dir keeps the temporary file
// on the same volume as the destination so the rename stays atomic, and the
// destination is never removed before the replacement is in place, so a failed
// or interrupted write cannot destroy an existing file.
func writeJSONOutput(path string, data []byte) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, "vaar-lint-*.json")
	if err != nil {
		return NewToolError("creating JSON output file failed", err)
	}
	tmpName := tmp.Name()
	// Clean up the temporary file on any early return. This is a no-op once the
	// rename below consumes it.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return NewToolError(fmt.Sprintf("writing JSON output to %s failed", path), err)
	}
	if err := tmp.Close(); err != nil {
		return NewToolError(fmt.Sprintf("writing JSON output to %s failed", path), err)
	}

	// os.Rename replaces an existing destination on all supported platforms
	// (Windows maps to MoveFileEx with MOVEFILE_REPLACE_EXISTING). On failure we
	// surface the error and leave any existing file untouched rather than
	// removing it first and risking a state with neither file present.
	if err := os.Rename(tmpName, path); err != nil {
		return NewToolError(fmt.Sprintf("writing JSON output to %s failed", path), err)
	}

	return nil
}

// hasTrailingSeparator reports whether path ends with a path separator, which
// signals a directory target rather than a file destination.
func hasTrailingSeparator(path string) bool {
	if path == "" {
		return false
	}
	last := path[len(path)-1]
	return last == '/' || last == os.PathSeparator
}
