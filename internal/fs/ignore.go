/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package fs

// ignoredDirectories keeps discovery out of generated, dependency and fixture
// trees that would add noise without changing lint behavior.
var ignoredDirectories = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"vendor":       {},
	"dist":         {},
	"build":        {},
	"testdata":     {},
}

func shouldSkipDir(name string) bool {
	_, ok := ignoredDirectories[name]
	return ok
}
