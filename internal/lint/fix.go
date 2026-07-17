/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

import (
	"bytes"
	"os"
	"sort"

	"github.com/envaar/vaar/internal/envfile"
)

// fixOrder lists the fixable rule IDs in the order their fix halves must
// compose to reproduce the historical whole-file normalization exactly. The
// order is load-bearing: line endings are normalized before trailing
// whitespace is trimmed, and blank runs are collapsed before the final blank
// lines are removed. TestFixDataMatchesLegacyNormalize pins this against a
// frozen copy of the original Normalize, so the order cannot drift silently.
var fixOrder = []string{
	"bom-character",
	"line-ending",
	"trailing-whitespace",
	"extra-blank-line",
	"ending-blank-line",
}

// composeFixes returns the fix halves of the fixable rules among the provided
// set. The fixable rules named in fixOrder come first, in that canonical order,
// so the built-in pipeline reproduces the historical normalization. Any other
// fixable rule — an external rule not named in fixOrder — is appended after that
// prefix in ascending-ID order, so its fix actually runs instead of silently
// doing nothing despite the rule showing FIXABLE in --list-rules. Rules without
// a fix half contribute nothing, so the caller still controls which repairs run
// by choosing which rules to pass in.
func composeFixes(rules []Rule) []func([]byte) []byte {
	byID := make(map[string]FixableRule, len(rules))
	for _, rule := range rules {
		if fixable, ok := rule.(FixableRule); ok {
			byID[rule.ID()] = fixable
		}
	}

	inOrder := make(map[string]bool, len(fixOrder))
	fixes := make([]func([]byte) []byte, 0, len(byID))
	for _, id := range fixOrder {
		inOrder[id] = true
		if fixable, ok := byID[id]; ok {
			fixes = append(fixes, fixable.Fix)
		}
	}

	extra := make([]string, 0, len(byID))
	for id := range byID {
		if !inOrder[id] {
			extra = append(extra, id)
		}
	}
	sort.Strings(extra)
	for _, id := range extra {
		fixes = append(fixes, byID[id].Fix)
	}
	return fixes
}

// FixData applies the fix halves of the given rules to data and returns the
// repaired bytes: the canonical fixOrder rules first, then any other fixable
// rule in ascending-ID order. Passing every built-in rule reproduces the
// historical whole-file normalization; passing a subset scopes the repair to
// those rules. It performs no I/O so tests can pin the composition directly.
func FixData(rules []Rule, data []byte) []byte {
	for _, fix := range composeFixes(rules) {
		data = fix(data)
	}
	return data
}

// ApplyFixes repairs each discovered dotenv file by composing the fix halves of
// the provided rules and reports whether any file changed on disk.
func ApplyFixes(rules []Rule, paths []string) (bool, error) {
	fixes := composeFixes(rules)
	changed := false
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}

		fixed := data
		for _, fix := range fixes {
			fixed = fix(fixed)
		}
		if bytes.Equal(fixed, data) {
			continue
		}

		if err := envfile.Write(path, fixed); err != nil {
			return false, err
		}
		changed = true
	}

	return changed, nil
}
