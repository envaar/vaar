/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteJSONOutputLeavesDestinationIntactOnFailure exercises the core safety
// guarantee behind the removed remove-then-rename fallback: when the final
// rename fails, writeJSONOutput must return an error and must NOT delete or
// truncate whatever already exists at the destination.
//
// A non-empty directory is used as the destination because renaming a file onto
// it fails on every platform, giving a portable way to force the failure path.
func TestWriteJSONOutputLeavesDestinationIntactOnFailure(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	if err := os.Mkdir(dest, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	inner := filepath.Join(dest, "keep.txt")
	if err := os.WriteFile(inner, []byte("IMPORTANT"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := writeJSONOutput(dest, []byte("{\"findings\":[]}\n")); err == nil {
		t.Fatal("expected writeJSONOutput to fail when the destination is a non-empty directory")
	}

	// The destination and its contents must be untouched.
	info, err := os.Stat(dest)
	if err != nil || !info.IsDir() {
		t.Fatalf("destination directory was destroyed (err=%v)", err)
	}
	data, err := os.ReadFile(inner)
	if err != nil {
		t.Fatalf("destination contents were destroyed: %v", err)
	}
	if string(data) != "IMPORTANT" {
		t.Fatalf("destination contents were modified: got %q", string(data))
	}

	// The temporary file must have been cleaned up, leaving only "dest" behind.
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "dest" {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected only dest/ to remain, got %v", names)
	}
}
