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
	mustWrite(filepath.Join(root, "app", ".env.prod"), "PROD=value\n")
	mustWrite(filepath.Join(root, "app", ".env.production.local"), "LOCAL=value\n")
	mustWrite(filepath.Join(root, "examples", "broken", ".env.example"), "BROKEN=value\n")
	mustWrite(filepath.Join(root, "dist", ".env.local"), "IGNORED=value\n")
	mustWrite(filepath.Join(root, "node_modules", ".env.production"), "IGNORED=value\n")
	mustWrite(filepath.Join(root, "testdata", ".env"), "IGNORED=value\n")
	mustWrite(filepath.Join(root, ".env."), "IGNORED=value\n")
	mustWrite(filepath.Join(root, ".environment"), "IGNORED=value\n")
	mustWrite(filepath.Join(root, ".envrc"), "IGNORED=value\n")
	mustWrite(filepath.Join(root, "my.env"), "IGNORED=value\n")

	files, err := fs.Discover(root)
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}

	if got, want := len(files), 5; got != want {
		t.Fatalf("unexpected file count: got %d want %d", got, want)
	}

	want := []string{
		filepath.Join(root, ".env"),
		filepath.Join(root, "app", ".env.example"),
		filepath.Join(root, "app", ".env.prod"),
		filepath.Join(root, "app", ".env.production.local"),
		filepath.Join(root, "examples", "broken", ".env.example"),
	}
	for i, wantPath := range want {
		if files[i] != wantPath {
			t.Fatalf("unexpected discovery result at %d: got %q want %q", i, files[i], wantPath)
		}
	}
}

func TestDiscoverSingleFileRecognizesDotenvFilenamePattern(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: ".env"},
		{name: ".env.local"},
		{name: ".env.prod"},
		{name: ".env.temp"},
		{name: ".env.preview-local"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tc.name)
			if err := os.WriteFile(path, []byte("KEY=value\n"), 0o644); err != nil {
				t.Fatalf("write failed: %v", err)
			}

			files, err := fs.Discover(path)
			if err != nil {
				t.Fatalf("discover failed: %v", err)
			}
			if len(files) != 1 || files[0] != path {
				t.Fatalf("unexpected discovery result: %#v", files)
			}
		})
	}
}
