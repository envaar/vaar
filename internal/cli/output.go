/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// validateOutputDestination rejects --output values that name a directory: a
// trailing path separator or a path that already exists as a directory. Without
// this guard the write path below creates a temporary file, fails to rename it
// onto the directory and then removes the directory, destroying it.
func validateOutputDestination(output string) error {
	if hasTrailingSeparator(output) {
		return outputDirectoryError(output)
	}
	if info, err := os.Stat(output); err == nil && info.IsDir() {
		return outputDirectoryError(output)
	}
	return nil
}

// writeJSONOutput writes data to path via a temporary file in the destination
// directory followed by an atomic rename. filepath.Dir keeps the temporary file
// on the same volume as the destination so the rename stays atomic. If the
// destination is a regular file on Windows and a direct rename fails, the file
// is removed and the rename is retried; existing directories are rejected and
// never removed.
func writeJSONOutput(path string, data []byte) error {
	dir := outputTempDir(path)

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
		return outputWriteError(path, err)
	}

	return finalizeJSONOutput(tmpName, path)
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

// outputTempDir keeps temporary JSON output alongside the final destination so
// the eventual rename stays on the same volume.
func outputTempDir(path string) string {
	return filepath.Dir(path)
}

func finalizeJSONOutput(tmpName, path string) error {
	renameErr := os.Rename(tmpName, path)
	if renameErr == nil {
		return nil
	}

	info, statErr := os.Stat(path)
	if statErr == nil && info.IsDir() {
		return outputDirectoryError(path)
	}
	if runtime.GOOS == "windows" && statErr == nil {
		if err := os.Remove(path); err != nil {
			return outputWriteError(path, err)
		}
		if err := os.Rename(tmpName, path); err != nil {
			return outputWriteError(path, err)
		}
		return nil
	}

	return outputWriteError(path, renameErr)
}

func outputDirectoryError(path string) error {
	return NewToolError(fmt.Sprintf("cannot write lint output to %q: the path is a directory", path), nil)
}

func outputWriteError(path string, err error) error {
	return NewToolError(fmt.Sprintf("writing JSON output to %s failed", path), err)
}
