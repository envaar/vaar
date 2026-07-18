/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package fs finds dotenv files in a repository and skips generated, vendored,
// and fixture directories that should not influence lint output.
package fs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Discover walks root, returns dotenv files in stable order and skips build,
// dependency, VCS and fixture directories that should not be linted.
func Discover(root string) ([]string, error) {
	if root == "" {
		root = "."
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if isDotenvFile(filepath.Base(absRoot)) {
			return []string{absRoot}, nil
		}
		return nil, errors.New("root is not a dotenv file or directory")
	}

	var files []string
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if path != absRoot && shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isDotenvFile(d.Name()) {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func isDotenvFile(name string) bool {
	return name == ".env" || (strings.HasPrefix(name, ".env.") && name != ".env.")
}
