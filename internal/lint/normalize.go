/*
Copyright © 2026 envaar
SPDX-License-Identifier: Apache-2.0
*/

package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/envaar/vaar/internal/envfile"
)

// normalizePipeline is the blunt whole-file formatting normalization exposed by
// the normalize command, in the canonical order the fix pipeline documents:
//
//	StripBOM -> NormalizeLineEndings -> TrimTrailingWhitespace ->
//	CollapseBlankLines -> TrimFinalBlankLines
//
// Unlike the rule fix halves behind lint --fix, every transform here runs
// unconditionally: line endings are forced to LF even when a file is uniform
// CRLF or uniform lone CR, so the pass does not depend on any rule reporting a
// finding.
var normalizePipeline = []func([]byte) []byte{
	envfile.StripBOM,
	envfile.NormalizeLineEndings,
	envfile.TrimTrailingWhitespace,
	envfile.CollapseBlankLines,
	envfile.TrimFinalBlankLines,
}

// NormalizeData applies the blunt normalization pipeline to data and returns
// the normalized bytes. It performs no I/O so tests can pin the pipeline
// directly.
func NormalizeData(data []byte) []byte {
	for _, transform := range normalizePipeline {
		data = transform(data)
	}
	return data
}

// Normalize discovers dotenv files for the given scope and rewrites each one
// whose bytes change under the blunt normalization pipeline, returning the
// rewritten paths relative to the root in discovery order. A file that is
// already normalized is not rewritten, so repeated runs are stable. Like
// ApplyFixes, it reports the first read or write failure and stops.
func Normalize(opts Options) ([]string, error) {
	if opts.Root == "" {
		opts.Root = "."
	}

	absRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, fmt.Errorf("resolve root %q: %w", opts.Root, err)
	}

	paths, err := discoverPaths(absRoot, opts.Root, opts.Target, opts.TargetDir)
	if err != nil {
		return nil, err
	}

	normalized := make([]string, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		fixed := NormalizeData(data)
		if bytes.Equal(fixed, data) {
			continue
		}

		if err := envfile.Write(path, fixed); err != nil {
			return nil, err
		}
		normalized = append(normalized, displayPath(absRoot, path))
	}

	return normalized, nil
}
