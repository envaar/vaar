/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

// Package diff compares dotenv key inventories without exposing values.
package diff

import (
	"fmt"
	"sort"

	"github.com/envaar/vaar/internal/envfile"
)

// Result describes the key-presence difference between two dotenv files.
type Result struct {
	// Left is the label or path for the first compared file.
	Left string
	// Right is the label or path for the second compared file.
	Right string
	// MissingFromLeft contains keys present in Right but absent from Left.
	MissingFromLeft []string
	// MissingFromRight contains keys present in Left but absent from Right.
	MissingFromRight []string
}

// HasDifferences reports whether either file is missing keys from the other.
func (r Result) HasDifferences() bool {
	return len(r.MissingFromLeft) > 0 || len(r.MissingFromRight) > 0
}

// Compare parses two dotenv inputs and compares assignment key presence only.
func Compare(leftPath string, leftData []byte, rightPath string, rightData []byte) (Result, error) {
	left, err := envfile.Parse(leftPath, leftData)
	if err != nil {
		return Result{}, fmt.Errorf("parse %q: %w", leftPath, err)
	}

	right, err := envfile.Parse(rightPath, rightData)
	if err != nil {
		return Result{}, fmt.Errorf("parse %q: %w", rightPath, err)
	}

	return CompareFiles(left, right), nil
}

// CompareFiles compares key presence between two already parsed dotenv files.
func CompareFiles(left envfile.File, right envfile.File) Result {
	leftKeys := keySet(left)
	rightKeys := keySet(right)

	return Result{
		Left:             left.Path,
		Right:            right.Path,
		MissingFromLeft:  missingKeys(leftKeys, rightKeys),
		MissingFromRight: missingKeys(rightKeys, leftKeys),
	}
}

func keySet(file envfile.File) map[string]struct{} {
	keys := make(map[string]struct{})
	for _, line := range file.Lines {
		if line.HasAssignment && line.HasKey {
			keys[line.Key] = struct{}{}
		}
	}
	return keys
}

func missingKeys(have map[string]struct{}, want map[string]struct{}) []string {
	missing := make([]string, 0)
	for key := range want {
		if _, ok := have[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}
