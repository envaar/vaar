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
		path, err := statScopeArg(root, target, "--target", false)
		if err != nil {
			return nil, err
		}
		return []string{path}, nil
	}

	if targetDir != "" {
		path, err := statScopeArg(root, targetDir, "--target-dir", true)
		if err != nil {
			return nil, err
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

// resolvePath turns a user-supplied path into an absolute path. Relative
// inputs are joined to root, while absolute inputs are cleaned and returned
// unchanged.
func resolvePath(root, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(root, path)
}

// statScopeArg resolves a scope input relative to root, checks whether it
// exists, verifies whether it is a file or directory, and returns the resolved
// path when the input is usable.
func statScopeArg(root, arg, flag string, wantDir bool) (string, error) {
	path := resolvePath(root, arg)
	info, err := os.Stat(path)
	switch {
	case err == nil:
		if info.IsDir() != wantDir {
			kind := "a file"
			if wantDir {
				kind = "a directory"
			}
			return "", fmt.Errorf("%s must point to %s: %s", flag, kind, arg)
		}
		return path, nil
	case os.IsNotExist(err):
		return "", fmt.Errorf("%s path does not exist: %s", flag, arg)
	default:
		return "", fmt.Errorf("%s path cannot be read: %s: %w", flag, arg, err)
	}
}
