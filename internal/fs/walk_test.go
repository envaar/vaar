/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package fs_test exercises directory discovery against real on-disk fixtures,
// including skipped build trees and single-file roots.
package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/envaar/vaar/internal/fs"
)

func TestDiscoverFindsDotenvFilesAndSkipsBuildDirs(t *testing.T) {
	root := t.TempDir()

	mustWrite := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	mustWrite(filepath.Join(root, ".env"), "KEY=value\n")
	mustWrite(filepath.Join(root, "app", ".env.example"), "EXAMPLE=value\n")
	mustWrite(filepath.Join(root, "examples", "broken", ".env.example"), "BROKEN=value\n")
	mustWrite(filepath.Join(root, "dist", ".env.local"), "IGNORED=value\n")
	mustWrite(filepath.Join(root, "node_modules", ".env.production"), "IGNORED=value\n")
	mustWrite(filepath.Join(root, "testdata", ".env"), "IGNORED=value\n")

	files, err := fs.Discover(root)
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}

	if got, want := len(files), 3; got != want {
		t.Fatalf("unexpected file count: got %d want %d", got, want)
	}

	wantFirst := filepath.Join(root, ".env")
	wantSecond := filepath.Join(root, "app", ".env.example")
	wantThird := filepath.Join(root, "examples", "broken", ".env.example")
	if files[0] != wantFirst || files[1] != wantSecond || files[2] != wantThird {
		t.Fatalf("unexpected discovery result: %#v", files)
	}
}
