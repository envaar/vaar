/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package scope resolves which dotenv files participate in a command.
package scope

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/envaar/vaar/internal/fs"
)

// Options describe the input scope for a command.
type Options struct {
	// Root is the repository root to scan.
	Root string
	// Target limits a run to one explicit file path.
	Target string
	// TargetDir limits a run to dotenv files discovered under one directory.
	TargetDir string
}

// Selection is the resolved set of input files for one command run.
type Selection struct {
	// Root is the absolute repository root used to resolve relative paths.
	Root string
	// RootLabel preserves the original root label for diagnostics.
	RootLabel string
	// Paths contains the ordered absolute input paths.
	Paths []string
}

// Resolve turns command scope options into a stable ordered selection.
func Resolve(opts Options) (Selection, error) {
	rootLabel := opts.Root
	if rootLabel == "" {
		rootLabel = "."
	}

	absRoot, err := filepath.Abs(rootLabel)
	if err != nil {
		return Selection{}, fmt.Errorf("resolve root %q: %w", rootLabel, err)
	}
	absRoot = canonicalizeExisting(absRoot)

	paths, err := discoverPaths(absRoot, rootLabel, opts.Target, opts.TargetDir)
	if err != nil {
		return Selection{}, err
	}

	return Selection{
		Root:      absRoot,
		RootLabel: rootLabel,
		Paths:     paths,
	}, nil
}

// DisplayPath keeps a path relative to the selected root when possible.
func (s Selection) DisplayPath(path string) string {
	rel, err := filepath.Rel(s.Root, path)
	if err != nil {
		return path
	}
	return rel
}

func discoverPaths(root, rootLabel, target, targetDir string) ([]string, error) {
	if target != "" && targetDir != "" {
		return nil, fmt.Errorf("--target and --target-dir cannot be used together")
	}

	if target != "" {
		path := resolvePath(root, target)
		info, err := os.Stat(path)
		switch {
		case err == nil:
			if info.IsDir() {
				return nil, fmt.Errorf("--target must point to a file: %s", target)
			}
			return []string{canonicalizeExisting(path)}, nil
		case os.IsNotExist(err):
			return nil, fmt.Errorf("--target path does not exist: %s", target)
		default:
			return nil, fmt.Errorf("--target path cannot be read: %s: %w", target, err)
		}
	}

	if targetDir != "" {
		path := resolvePath(root, targetDir)
		info, err := os.Stat(path)
		switch {
		case err == nil:
			if !info.IsDir() {
				return nil, fmt.Errorf("--target-dir must point to a directory: %s", targetDir)
			}
			path = canonicalizeExisting(path)
		case os.IsNotExist(err):
			return nil, fmt.Errorf("--target-dir path does not exist: %s", targetDir)
		default:
			return nil, fmt.Errorf("--target-dir path cannot be read: %s: %w", targetDir, err)
		}

		paths, err := fs.Discover(path)
		if err != nil {
			return nil, fmt.Errorf("discovering files under %q: %w", targetDir, err)
		}
		return paths, nil
	}

	paths, err := fs.Discover(root)
	if err != nil {
		return nil, fmt.Errorf("discovering files under %q: %w", rootLabel, err)
	}
	return paths, nil
}

// resolvePath turns a user-supplied relative or absolute path into an
// absolute path anchored to root.
func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}

// canonicalizeExisting resolves symlinks for a path when possible. If the
// path cannot be resolved, the original path is preserved.
func canonicalizeExisting(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}
